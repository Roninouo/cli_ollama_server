param(
  [switch] $DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Here = Split-Path -Parent $MyInvocation.MyCommand.Path

function Normalize-Path([string] $p) {
  return ([IO.Path]::GetFullPath($p)).TrimEnd('\\')
}

$Target = Normalize-Path $Here
$CurrentUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($null -eq $CurrentUserPath) { $CurrentUserPath = '' }

$Parts = @()
if ($CurrentUserPath.Trim().Length -gt 0) {
  $Parts = $CurrentUserPath.Split(';') | Where-Object { $_ -and $_.Trim().Length -gt 0 }
}

$Kept = New-Object System.Collections.Generic.List[string]
foreach ($p in $Parts) {
  $keep = $true
  try {
    if ((Normalize-Path $p) -ieq $Target) { $keep = $false }
  } catch {
    # keep invalid path segments as-is
  }
  if ($keep) { $Kept.Add($p) }
}

$NewUserPath = ($Kept -join ';')

if ($DryRun) {
  Write-Output 'Dry run (no changes made).'
  Write-Output ("Would remove from User PATH: {0}" -f $Target)
  exit 0
}

[Environment]::SetEnvironmentVariable('Path', $NewUserPath, 'User')
$env:Path = $NewUserPath

Write-Output ("Removed from User PATH: {0}" -f $Target)
Write-Output 'Open a new terminal to pick up the change everywhere.'
