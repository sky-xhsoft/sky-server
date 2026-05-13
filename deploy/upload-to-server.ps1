# Sky-Server 自动上传脚本 - PowerShell
# 使用环境变量配置服务器连接信息:
#   $env:SKY_DEPLOY_HOST     - 目标服务器地址
#   $env:SKY_DEPLOY_USER     - SSH 用户名
#   $env:SKY_DEPLOY_PASSWORD - SSH 密码 (或使用 SSH Key)

$server = $env:SKY_DEPLOY_HOST
$username = $env:SKY_DEPLOY_USER
$password = $env:SKY_DEPLOY_PASSWORD

if (-not $server -or -not $username) {
    Write-Error "请设置环境变量 SKY_DEPLOY_HOST 和 SKY_DEPLOY_USER"
    exit 1
}

$files = @(
    "bin\sky-server",
    "deploy\init-server.sh",
    "deploy\deploy.sh",
    "deploy\sky-server.service",
    "configs\config.example.yaml"
)

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "   Sky-Server 文件上传到服务器"
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Server: $server"
Write-Host "User: $username"
Write-Host ""

foreach ($file in $files) {
    Write-Host "Uploading $file..." -ForegroundColor Yellow
    $target = "${username}@${server}:/tmp/"
    & scp -o StrictHostKeyChecking=no $file $target
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to upload $file" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "   所有文件上传完成!"
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""
Write-Host "下一步："
Write-Host "1. SSH 登录服务器: ssh root@$server"
Write-Host "2. 执行初始化: cd /tmp && chmod +x init-server.sh && ./init-server.sh"
Write-Host ""
