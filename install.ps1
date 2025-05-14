<#  jellyfaas CLI one-shot installer  --------------------------------------
    • No admin/UAC required
    • Installs under %LOCALAPPDATA%\jellyfaas\bin
    • Adds that folder to the *user* PATH (persists for new shells)
---------------------------------------------------------------------------#>

$ErrorActionPreference = 'Stop'          # fail fast

# ---- configurable bits ----------------------------------------------------
$Version    = 'v1.0.0'
$Arch       = 'win32'
$ZipUrl     = "https://github.com/Platform48/jellyfaas_cli/releases/download/$Version/jellyfaas-$Version-$Arch.zip"

# ---- derived paths --------------------------------------------------------
$TempFile   = Join-Path $env:TEMP "jellyfaas_$Version.zip"
$RootDir    = Join-Path $env:LOCALAPPDATA "jellyfaas"
$BinDir     = Join-Path $RootDir "bin"

# ---- download -------------------------------------------------------------
Write-Host  "⤵  Downloading $ZipUrl …"
Invoke-WebRequest -Uri $ZipUrl -OutFile $TempFile

# ---- unpack ---------------------------------------------------------------
Write-Host  "📦  Extracting to $RootDir"
New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
Expand-Archive -Path $TempFile -DestinationPath $RootDir -Force

# Some GitHub zips include a top-level folder; move the EXE(s) into bin
Get-ChildItem -Path $RootDir -Filter 'jellyfaas*.exe' -Recurse |
    ForEach-Object { Move-Item $_.FullName -Destination $BinDir -Force }

# ---- PATH handling --------------------------------------------------------
$PathAlready = $env:PATH.Split(';') -contains $BinDir
if (-not $PathAlready) {
    # make it available *now* …
    $env:PATH = "$BinDir;$env:PATH"
    # … and persist for *future* shells
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (-not $userPath) { $userPath = '' }
    if (-not ($userPath.Split(';') -contains $BinDir)) {
        [Environment]::SetEnvironmentVariable('Path', "$BinDir;$userPath", 'User')
        $Persisted = $true
    }
}

# ---- cleanup --------------------------------------------------------------
Remove-Item $TempFile -Force

# ---- done -----------------------------------------------------------------
Write-Host "`n✅  jellyfaas CLI installed in $BinDir"
if ($Persisted) {
    Write-Host "🔄  Restart PowerShell (or run 'refreshenv' if you have it) so new PATH is picked up."
} else {
    Write-Host "ℹ   PATH was already configured."
}
