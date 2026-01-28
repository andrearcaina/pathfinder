$ErrorActionPreference = "Stop" # stops the script if any error occurs (Write-Error)

$Repo = "andrearcaina/pathfinder"
$BinName = "pathfinder"
$BinNameExe = "$BinName.exe"

# get the latest tag from GitHub
Write-Host "Fetching latest release tag..."
$LatestReleaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
try {
    $Response = Invoke-RestMethod -Uri $LatestReleaseUrl -Method Get
    $Tag = $Response.tag_name
} catch {
    Write-Error "Failed to fetch latest release tag."
    exit 1
}

# add "v" if missing (though github usually includes it, just in case logic differs)
if (-not $Tag.StartsWith("v")) {
    $Tag = "v$Tag"
}

# detect architecture
$ArchInput = $env:PROCESSOR_ARCHITECTURE

if ($ArchInput -eq "AMD64") {
    $Arch = "x86_64"
} elseif ($ArchInput -eq "ARM64") {
    $Arch = "arm64"
} elseif ($ArchInput -eq "x86") {
    if ([System.Environment]::Is64BitOperatingSystem) {
        $Arch = "x86_64"
    } else {
        $Arch = "i386"
    }
} else {
    Write-Error "Unsupported architecture: $ArchInput"
    exit 1
}

$Os = "Windows"
$Ext = "zip"
$DownloadName = "${BinName}_${Os}_${Arch}.${Ext}"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$DownloadName"

Write-Host "Downloading $Tag for Windows ($Arch)..."
$TempZip = Join-Path $env:TEMP "$BinName.zip"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip
} catch {
    Write-Error "Failed to download release from $DownloadUrl"
    exit 1
}

# install directory (this is where the binary will be installed)
# will be installed in $USERPROFILE/.pathfinder/bin
$InstallDir = Join-Path $env:USERPROFILE ".pathfinder/bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# extract (extract the zip file downloaded from GitHub)
Expand-Archive -Path $TempZip -DestinationPath $env:TEMP -Force
$ExtractedBin = Join-Path $env:TEMP $BinNameExe

if (Test-Path $ExtractedBin) {
    Move-Item -Path $ExtractedBin -Destination (Join-Path $InstallDir $BinNameExe) -Force
    Write-Host "Installed to $InstallDir\$BinNameExe"
} else {
    Write-Error "Could not find $BinNameExe after extraction."
    exit 1
}
# then cleanup (remove temp zip)
Remove-Item $TempZip -Force

# check USER PATH and add pathfinder bin directory if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to User PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", [EnvironmentVariableTarget]::User)
    $env:PATH = "$env:PATH;$InstallDir"
    Write-Host "Added to PATH. You may need to restart your shell for changes to take full effect."
} else {
    Write-Host "$InstallDir is already in your User PATH."
}

Write-Host "Installation complete."
