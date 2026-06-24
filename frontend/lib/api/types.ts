// 通用 API 类型。
export interface ListResponse<T> {
  items: T[];
  total?: number;
}
