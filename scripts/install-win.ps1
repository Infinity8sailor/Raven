$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$appDir = (Resolve-Path "$scriptDir\..").Path
$exePath = Join-Path $appDir "raven.exe"

$action = New-ScheduledTaskAction -Execute $exePath -WorkingDirectory $appDir
$trigger = New-ScheduledTaskTrigger -AtLogOn
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "RavenApp" -Description "Starts Raven minimal backend at logon" -RunLevel Highest
Write-Host "Raven scheduled task installed on Windows for $exePath."
