// 文档解析 worker：上传后异步执行「解析 → 存解析件 → 结构化提取」管道。
// Excel/CSV 走本地直读（快、准）；PDF/图片/Word 走 docling OCR + GLM 定向提取。
// 状态流转：uploaded → parsing → parsed / failed。
package document

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/docling"
	"github.com/ptis/backend/internal/storage"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	repo      *db.DocumentRepository
	docling   *docling.Client
	store     *storage.ObjectStore
	importers *Importers    // 可选：解析后自动入库
	sem       chan struct{} // 并发上限：OCR 调外部视觉 API，避免挤爆
}

func NewWorker(repo *db.DocumentRepository, dl *docling.Client, store *storage.ObjectStore,
	importers *Importers) *Worker {
	return &Worker{repo: repo, docling: dl, store: store, importers: importers, sem: make(chan struct{}, 2)}
}

// SourceKind 按文件名后缀归类解析路径。
func SourceKind(filename string) string {
	switch strings.ToLower(path.Ext(filename)) {
	case ".xlsx", ".xlsm", ".xls":
		return "excel"
	case ".csv", ".txt":
		return "csv"
	case ".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".webp":
		return "image"
	case ".docx", ".doc":
		return "word"
	default:
		return "pdf"
	}
}

// OriginalKey / ParsedKey 对象存储 key 约定。
func OriginalKey(orgID, docID, filename string) string {
	if orgID == "" {
		orgID = "shared"
	}
	return fmt.Sprintf("documents/%s/%s/%s", orgID, docID, path.Base(filename))
}

func ParsedKey(orgID, docID string) string {
	if orgID == "" {
		orgID = "shared"
	}
	return fmt.Sprintf("documents/%s/%s/parsed.md", orgID, docID)
}

// Enqueue 异步解析。content 为上传时的原件字节；传 nil 表示从对象存储取回（重新解析）。
func (w *Worker) Enqueue(docID uuid.UUID, content []byte) {
	go func() {
		w.sem <- struct{}{}
		defer func() { <-w.sem }()
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Str("doc", docID.String()).Msg("文档解析 panic")
				_ = w.repo.SetStatus(context.Background(), docID, "failed", fmt.Sprintf("解析异常: %v", r))
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		if err := w.process(ctx, docID, content); err != nil {
			log.Error().Err(err).Str("doc", docID.String()).Msg("文档解析失败")
			_ = w.repo.SetStatus(context.Background(), docID, "failed", err.Error())
		}
	}()
}

func (w *Worker) process(ctx context.Context, docID uuid.UUID, content []byte) error {
	doc, err := w.repo.GetByIDInternal(ctx, docID)
	if err != nil || doc == nil {
		return fmt.Errorf("读取文档记录失败: %v", err)
	}
	if err := w.repo.SetStatus(ctx, docID, "parsing", ""); err != nil {
		return err
	}

	// 重新解析：从对象存储取回原件
	if content == nil {
		if w.store.Enabled() && doc.OriginalObjectKey != nil && *doc.OriginalObjectKey != "" {
			rc, err := w.store.Get(ctx, *doc.OriginalObjectKey)
			if err != nil {
				return fmt.Errorf("取回原件失败: %w", err)
			}
			content, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("读取原件失败: %w", err)
			}
		} else {
			return fmt.Errorf("原件未存档，无法重新解析")
		}
	}

	var (
		docType  string
		text     string
		tables   json.RawMessage
		entities json.RawMessage
		summary  string
		pages    int
		exts     []db.ExtractionInput
	)

	switch doc.SourceKind {
	case "excel", "csv":
		res, err := parseTabular(doc.Filename, content)
		if err != nil {
			return err
		}
		docType, text, tables, summary = res.DocType, res.TextContent, res.Tables, res.Summary
		pages = res.SheetCount
		exts = res.Extractions
		entities = json.RawMessage(`{}`)

	default: // pdf / image / word → docling OCR
		res, err := w.docling.ParseDoc(ctx, doc.Filename, content, "auto")
		if err != nil {
			return err
		}
		docType, text, summary, pages = res.DocType, res.TextContent, res.Summary, res.PageCount
		tables, _ = json.Marshal(res.Tables)
		entities, _ = json.Marshal(res.Entities)

		// GLM 定向结构化提取；失败不阻断（正则实体仍可看）
		fields, err := w.docling.Extract(ctx, text, docType)
		if err != nil {
			log.Warn().Err(err).Str("doc", docID.String()).Msg("结构化提取失败，仅保留正则实体")
		}
		for _, f := range fields {
			exts = append(exts, db.ExtractionInput{
				GroupNo: 0, FieldKey: f.Key, FieldLabel: f.Label,
				ValueText: f.Value, Unit: f.Unit, Confidence: f.Confidence,
				Source: "glm",
			})
		}

		// PDF 中的表格 → 复用 Excel 模板识别，产出行级字段（支持批量入库）。
		// 取第一个命中业务模板的表格；OCR 表格置信度按 0.8 计（提示人工核对）。
		for _, tb := range res.Tables {
			mdStr, _ := tb["markdown"].(string)
			if mdStr == "" {
				continue
			}
			rows := markdownTableToRows(mdStr)
			if len(rows) < 2 {
				continue
			}
			tType, tExts := detectTemplate(rows)
			if len(tExts) == 0 {
				continue
			}
			for i := range tExts {
				tExts[i].Source = "glm"
				tExts[i].Confidence = 0.8
			}
			exts = append(exts, tExts...)
			if docType == "其他" || docType == "" {
				docType = tType
			}
			break
		}
	}

	// 解析件落对象存储（MinIO 未配置时跳过，仅存库内文本）
	parsedKey := ""
	if w.store.Enabled() {
		org := ""
		if doc.OrgID != nil {
			org = *doc.OrgID
		}
		parsedKey = ParsedKey(org, doc.ID)
		if err := w.store.Put(ctx, parsedKey, bytes.NewReader([]byte(text)),
			int64(len(text)), "text/markdown; charset=utf-8"); err != nil {
			log.Warn().Err(err).Msg("解析件写入对象存储失败（仅存库内文本）")
			parsedKey = ""
		}
	}

	if err := w.repo.SetParsed(ctx, docID, docType, pages, text, tables, entities, summary, parsedKey); err != nil {
		return fmt.Errorf("写回解析结果失败: %w", err)
	}
	if err := w.repo.ReplaceExtractions(ctx, docID, exts); err != nil {
		return fmt.Errorf("写入提取字段失败: %w", err)
	}
	log.Info().Str("doc", docID.String()).Str("type", docType).
		Int("fields", len(exts)).Msg("文档解析完成")

	// 解析完成 → 尝试自动入库（Excel 直接入；GLM 提取需置信度达标）
	w.tryAutoApply(ctx, doc, docType, exts)

	return nil
}

// tryAutoApply 解析后自动入库：
//   - Excel/CSV（source=excel, confidence=1.0）→ 直接入库
//   - GLM 提取（source=glm）→ 最低置信度 ≥ 0.85 才自动入库
//   - 无明确入库目标 → 跳过
//   - 失败不阻断（文档已 parsed，用户可手动入库）
func (w *Worker) tryAutoApply(ctx context.Context, doc *db.Document, docType string, exts []db.ExtractionInput) {
	if w.importers == nil || len(exts) == 0 {
		return
	}
	target := SuggestTarget(docType)
	if target == "" {
		return
	}

	// 计算最低置信度，判断是否有 GLM 来源字段
	minConf := 1.0
	hasGlm := false
	for _, e := range exts {
		if e.Source == "glm" {
			hasGlm = true
			if e.Confidence < minConf {
				minConf = e.Confidence
			}
		}
	}
	// 风险分级：高风险目标（建客户/合同）一律人工确认；GLM 提取置信度门槛 0.85。
	// 命中则不自动写库，留到文档详情由人核对后「一键归入」（那一步即人工确认）。
	if NeedsReview(target, hasGlm, minConf) {
		log.Info().Str("doc", doc.ID).Str("type", docType).Str("target", target).
			Float64("min_confidence", minConf).Bool("high_risk", reviewTargets[target]).
			Msg("需人工确认，跳过自动入库（在文档详情核对后一键归入）")
		return
	}

	// 构造提取字段指针列表（Importers.Apply 接收 []*DocumentExtraction）
	extPtrs := make([]*db.DocumentExtraction, len(exts))
	for i := range exts {
		val := exts[i].ValueText
		unit := exts[i].Unit
		conf := exts[i].Confidence
		extPtrs[i] = &db.DocumentExtraction{
			GroupNo:    exts[i].GroupNo,
			FieldKey:   exts[i].FieldKey,
			FieldLabel: exts[i].FieldLabel,
			ValueText:  &val,
			Unit:       &unit,
			Confidence: &conf,
			Source:     exts[i].Source,
		}
	}

	// 入库上下文带文档归属省
	applyCtx := ctx
	if doc.OrgID != nil && *doc.OrgID != "" {
		applyCtx = db.WithOrg(applyCtx, *doc.OrgID)
	}

	applied, detail, err := w.importers.Apply(applyCtx, target, doc.ID, extPtrs)
	if err != nil {
		log.Warn().Err(err).Str("doc", doc.ID).Msg("自动入库失败（待人工确认）")
		return
	}

	// 记录入库审计（applied_by=nil 表示系统自动）
	docID, _ := uuid.Parse(doc.ID)
	_ = w.repo.InsertApply(ctx, docID, target, applied, detail, nil)
	_ = w.repo.SetAutoApplied(ctx, docID, true)

	log.Info().Str("doc", doc.ID).Str("target", target).
		Int("applied", applied).Float64("min_confidence", minConf).
		Msg("文档自动入库成功")
}
