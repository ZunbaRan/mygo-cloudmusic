!macro customInit
  ; Check if MyGo Cloud Music is running before starting installation
  nsExec::ExecToStack 'cmd.exe /c "tasklist /NH /FI "IMAGENAME eq MyGo Cloud Music.exe" | findstr /I "MyGo""'
  Pop $0
  Pop $1
  StrCmp $0 "0" 0 not_running
    IfSilent do_kill
    MessageBox MB_OKCANCEL|MB_ICONQUESTION "MyGo Cloud Music is currently running.$\r$\n$\r$\nClick OK to close it and continue installing, or Cancel to exit." IDOK do_kill
    Abort
  do_kill:
    nsExec::Exec 'taskkill /F /T /IM "MyGo Cloud Music.exe"'
    Sleep 500
  not_running:
!macroend

!macro customUnInstall
  nsExec::Exec 'taskkill /F /T /IM "MyGo Cloud Music.exe"'
  Sleep 500
!macroend
