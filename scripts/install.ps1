# Requires -Version 5.0
param(
    [string]$InstallVersion = "latest"
)

function Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Blue }
function Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Err($msg) { Write-Host "[ERR] $msg" -ForegroundColor Red }

$REPO_OWNER = "anwerj"
$REPO_NAME = "youtube-uploader-mcp"
$BINARY_BASENAME = "youtube-uploader-mcp"
$WIN_DEFAULT_BIN = "$env:USERPROFILE\youtube-uploader-mcp"
$CLAUDE_CONFIG = "$env:APPDATA\Claude\claude_desktop_config.json"

function DetectArch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Err "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"; exit 1 }
    }
}

function GhReleaseUrl($arch, $version) {
    $asset = "$BINARY_BASENAME-windows-$arch.exe"
    if ($version -eq "latest") {
        return "https://github.com/$REPO_OWNER/$REPO_NAME/releases/latest/download/$asset"
    } else {
        return "https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$version/$asset"
    }
}

function ValidateClientSecret($file) {
    if (!(Test-Path $file)) { Err "Client secret file not found: $file"; exit 1 }
    $json = Get-Content $file | ConvertFrom-Json
    if (-not ($json.PSObject.Properties.Name -contains "installed" -or $json.PSObject.Properties.Name -contains "web")) {
        Err "client_secret.json does not look like a Google OAuth client secret."; exit 1
    }
    $cid = $json.installed.client_id
    if (-not $cid) { $cid = $json.web.client_id }
    if (-not $cid) { Err "No client_id found in client_secret.json"; exit 1 }
}

function FindClientSecret {
    $searchPaths = @(Get-Location, "$env:USERPROFILE\Downloads")
    foreach ($dir in $searchPaths) {
        $files = Get-ChildItem -Path $dir -Filter "client_secret*.apps.googleusercontent.com.json" -File -ErrorAction SilentlyContinue
        if ($files) { return $files[0].FullName }
    }
    Err "Could not find client_secret*.apps.googleusercontent.com.json in current directory or Downloads."; exit 1
}

function InstallBinary($arch, $version) {
    $url = GhReleaseUrl $arch $version
    $destDir = $WIN_DEFAULT_BIN
    if (!(Test-Path $destDir)) { New-Item -ItemType Directory -Path $destDir | Out-Null }
    $binaryPath = "$destDir\$BINARY_BASENAME-windows-$arch.exe"
    Info "Downloading $url"
    Invoke-WebRequest -Uri $url -OutFile $binaryPath
    Info "Installed binary: $binaryPath"
    return $binaryPath
}

function CopyClientSecret($src) {
    $cfgDir = "$env:APPDATA\youtube-uploader-mcp"
    if (!(Test-Path $cfgDir)) { New-Item -ItemType Directory -Path $cfgDir | Out-Null }
    $dst = "$cfgDir\$(Split-Path $src -Leaf)"
    Copy-Item $src $dst -Force
    Info "Client secret stored at: $dst"
    return $dst
}

function IntegrateClaude($cmd, $clientSecret) {
    $claudeCfg = $CLAUDE_CONFIG
    if (!(Test-Path $claudeCfg)) {
        $parent = Split-Path $claudeCfg -Parent
        if (!(Test-Path $parent)) { New-Item -ItemType Directory -Path $parent | Out-Null }
        Set-Content -Path $claudeCfg -Value "{}"
    }
    $json = Get-Content $claudeCfg | ConvertFrom-Json
    if (-not $json.mcpServers) { $json | Add-Member -MemberType NoteProperty -Name mcpServers -Value @{} }
    $json.mcpServers["youtube-uploader-mcp"] = @{ command = $cmd; args = @("-client_secret_file", $clientSecret) }
    $json | ConvertTo-Json -Depth 5 | Set-Content $claudeCfg
    Info "Claude config updated at: $claudeCfg"
}

function ShowVSCodeConfig($cmd, $clientSecret) {
    $config = @{ command = $cmd; args = @("-client_secret_file", $clientSecret); env = @{}; disabled = $false; autoApprove = @() }
    $json = $config | ConvertTo-Json -Depth 5
    Write-Host "\nAdd the following to your MCP config:" -ForegroundColor Blue
    Write-Host '"youtube-uploader-mcp": ' $json
}

# Main
Info "Detecting architecture..."
$arch = DetectArch

Info "Checking location of client secret file"
$clientSecretSrc = FindClientSecret
ValidateClientSecret $clientSecretSrc
$clientSecret = CopyClientSecret $clientSecretSrc

Info "Installing $REPO_OWNER/$REPO_NAME ($InstallVersion)"
$binaryPath = InstallBinary $arch $InstallVersion

$yn = Read-Host "Integrate with Claude Desktop now? [y/N]"
if ($yn -match '^[Yy]') { IntegrateClaude $binaryPath $clientSecret }

$yn2 = Read-Host "Show VS Code MCP config JSON now? [y/N]"
if ($yn2 -match '^[Yy]') { ShowVSCodeConfig $binaryPath $clientSecret }

Info "Setup complete."
