param([Parameter(Mandatory=$true)][string]$OutputDirectory)
$ErrorActionPreference='Stop'
& dotnet publish (Join-Path $PSScriptRoot 'EidPd.PinPrompt.csproj') -c Release -r win-x64 --self-contained true -p:RestoreLockedMode=true -o $OutputDirectory
if($LASTEXITCODE){throw 'WinUI PIN prompt build failed'}
if(!(Test-Path (Join-Path $OutputDirectory 'EidPd.PinPrompt.exe'))){throw 'WinUI PIN prompt output missing'}
$probe=Start-Process -FilePath (Join-Path $OutputDirectory 'EidPd.PinPrompt.exe') -ArgumentList '--check' -WindowStyle Hidden -PassThru
if(!$probe.WaitForExit(15000)){$probe.Kill();throw 'WinUI PIN prompt startup check timed out'}
if($probe.ExitCode -ne 0){throw "WinUI PIN prompt startup check failed: $($probe.ExitCode)"}
