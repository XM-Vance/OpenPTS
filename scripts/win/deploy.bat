@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion

:: =============================================================
:: OpenPTS Windows 11 一键部署脚本
:: 前置：Docker Desktop 已安装并运行
:: 用法：
::   deploy.bat          首次部署（构建+启动）
::   deploy.bat start    启动服务
::   deploy.bat stop     停止服务
::   deploy.bat restart  重启服务
::   deploy.bat status   查看状态
::   deploy.bat logs     查看日志
::   deploy.bat reset    重置（删数据，慎用）
:: =============================================================

title OpenPTS 部署工具

:: 项目目录（脚本在 scripts/win/ 下，需回到项目根）
set "PROJECT_DIR=%~dp0..\.."
cd /d "%PROJECT_DIR%"

:: 颜色
for /F %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "GREEN=%ESC%[92m"
set "RED=%ESC%[91m"
set "YELLOW=%ESC%[93m"
set "BLUE=%ESC%[94m"
set "RESET=%ESC%[0m"

echo.
echo =========================================
echo   OpenPTS 开放式电力交易系统
echo   Windows 部署工具 v1.0
echo =========================================
echo.

:: 检查 Docker Desktop
docker info >nul 2>&1
if errorlevel 1 (
    echo %RED%  ❌ Docker Desktop 未运行！%RESET%
    echo.
    echo   请先：
    echo   1. 安装 Docker Desktop: https://www.docker.com/products/docker-desktop/
    echo   2. 启动 Docker Desktop
    echo   3. 等待 Docker 引擎就绪后重新运行此脚本
    echo.
    pause
    exit /b 1
)
echo %GREEN%  ✅ Docker Desktop 已就绪%RESET%

:: 检查 docker compose 版本
docker compose version >nul 2>&1
if errorlevel 1 (
    echo %RED%  ❌ docker compose 不可用%RESET%
    pause
    exit /b 1
)
echo %GREEN%  ✅ Docker Compose 已就绪%RESET%

:: 使用 Windows 兼容的 compose 文件
set "DC=docker compose -f docker-compose.win.yml"

:: 命令处理
set "CMD=%1"
if "%CMD%"=="" set "CMD=deploy"

if "%CMD%"=="deploy" goto :deploy
if "%CMD%"=="start" goto :start
if "%CMD%"=="stop" goto :stop
if "%CMD%"=="restart" goto :restart
if "%CMD%"=="status" goto :status
if "%CMD%"=="logs" goto :logs
if "%CMD%"=="reset" goto :reset
if "%CMD%"=="seed" goto :seed

echo.
echo   用法: deploy.bat {deploy^|start^|stop^|restart^|status^|logs^|reset^|seed}
echo.
goto :end

:: ─── 首次部署 ───
:deploy
echo.
echo %BLUE%  📦 首次部署开始...%RESET%
echo.

:: 1. 检查 .env
if not exist ".env" (
    echo %YELLOW%  ⚠️  未找到 .env 文件，从模板创建...%RESET%
    copy ".env.example" ".env" >nul
    
    :: 生成随机密码
    for /f %%a in ('powershell -Command "[Convert]::ToBase64String((1..48 | ForEach-Object { Get-Random -Maximum 256 }) -as [byte[]])"') do set "JWT_SECRET=%%a"
    for /f %%a in ('powershell -Command "[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }) -as [byte[]])"') do set "PG_PASS=%%a"
    for /f %%a in ('powershell -Command "[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }) -as [byte[]])"') do set "MINIO_PASS=%%a"
    
    :: 替换密码
    powershell -Command "(Get-Content .env) -replace 'POSTGRES_PASSWORD=.*', 'POSTGRES_PASSWORD=%PG_PASS%' -replace 'JWT_SECRET=.*', 'JWT_SECRET=%JWT_SECRET%' -replace 'MINIO_SECRET_KEY=.*', 'MINIO_SECRET_KEY=%MINIO_PASS%' | Set-Content .env"
    
    echo %GREEN%  ✅ .env 已创建（随机密码已生成）%RESET%
) else (
    echo %GREEN%  ✅ .env 已存在%RESET%
)

:: 2. 构建镜像（容器内编译 Go/Node/Python，无需本地安装）
echo.
echo %BLUE%  🔨 构建 Docker 镜像（首次较慢，约5-10分钟）...%RESET%
echo.
%DC% build --parallel 2>&1
if errorlevel 1 (
    echo %RED%  ❌ 镜像构建失败%RESET%
    echo   查看上方错误信息，常见问题：
    echo   - 网络问题：配置 Docker 镜像加速器
    echo   - 内存不足：Docker Desktop 设置 ^>= 4GB 内存
    pause
    exit /b 1
)
echo %GREEN%  ✅ 镜像构建完成%RESET%

:: 3. 启动服务
echo.
echo %BLUE%  🚀 启动服务...%RESET%
%DC% up -d 2>&1
if errorlevel 1 (
    echo %RED%  ❌ 启动失败%RESET%
    pause
    exit /b 1
)

:: 4. 等待 PostgreSQL 就绪
echo.
echo %BLUE%  ⏳ 等待 PostgreSQL 就绪...%RESET%
set READY=0
for /L %%i in (1,1,30) do (
    if !READY! equ 0 (
        %DC% exec -T postgres pg_isready -U ptis >nul 2>&1
        if not errorlevel 1 (
            set READY=1
            echo %GREEN%  ✅ PostgreSQL 已就绪%RESET%
        ) else (
            echo   等待中... %%i/30
            timeout /t 2 /nobreak >nul
        )
    )
)
if %READY% equ 0 (
    echo %RED%  ❌ PostgreSQL 启动超时%RESET%
    pause
    exit /b 1
)

:: 5. 运行数据库迁移
echo.
echo %BLUE%  📊 运行数据库迁移...%RESET%
%DC% --profile migrate run --rm migrate up 2>&1
if errorlevel 1 (
    echo %YELLOW%  ⚠️  迁移可能已执行过，忽略%RESET%
) else (
    echo %GREEN%  ✅ 数据库迁移完成%RESET%
)

:: 6. 种子数据
echo.
set /p "SEED=是否导入演示数据？(Y/n): "
if /i "%SEED%"=="n" goto :skip_seed
echo %BLUE%  📥 导入演示数据...%RESET%
%DC% exec backend /app/seed 2>&1
echo %GREEN%  ✅ 演示数据已导入%RESET%
:skip_seed

echo.
echo =========================================
echo %GREEN%  🎉 部署完成！%RESET%
echo =========================================
echo.
echo   访问地址：
echo   🌐 系统界面: http://localhost
echo   📊 后端健康: http://localhost/api/v1/health
echo.
echo   管理命令：
echo   deploy.bat stop     停止服务
echo   deploy.bat start    启动服务
echo   deploy.bat status   查看状态
echo   deploy.bat logs     查看日志
echo   deploy.bat seed     导入演示数据
echo.
echo   数据持久化在 Docker 卷中，停止不会丢失。
echo =========================================
goto :end

:: ─── 启动 ───
:start
echo %BLUE%  🚀 启动服务...%RESET%
%DC% up -d 2>&1
echo %GREEN%  ✅ 服务已启动%RESET%
echo   访问: http://localhost
goto :end

:: ─── 停止 ───
:stop
echo %YELLOW%  ⏹️  停止服务...%RESET%
%DC% down 2>&1
echo %GREEN%  ✅ 服务已停止（数据保留）%RESET%
goto :end

:: ─── 重启 ───
:restart
echo %BLUE%  🔄 重启服务...%RESET%
%DC% down 2>&1
%DC% up -d 2>&1
echo %GREEN%  ✅ 服务已重启%RESET%
goto :end

:: ─── 状态 ───
:status
echo.
echo %BLUE%  📋 服务状态：%RESET%
echo.
%DC% ps 2>&1
echo.
echo   访问: http://localhost
goto :end

:: ─── 日志 ───
:logs
%DC% logs -f --tail=100 2>&1
goto :end

:: ─── 种子数据 ───
:seed
echo %BLUE%  📥 导入演示数据...%RESET%
%DC% exec backend /app/seed 2>&1
echo %GREEN%  ✅ 完成%RESET%
goto :end

:: ─── 重置 ───
:reset
echo.
echo %RED%  ⚠️  警告：此操作将删除所有数据！%RESET%
echo.
set /p "CONFIRM=确认重置？输入 YES 继续: "
if not "%CONFIRM%"=="YES" (
    echo 已取消。
    goto :end
)
%DC% down -v 2>&1
echo %GREEN%  ✅ 已重置（数据卷已删除）%RESET%
echo   运行 deploy.bat 重新部署
goto :end

:end
echo.
