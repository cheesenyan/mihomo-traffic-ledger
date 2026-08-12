#define MyAppName "Clash 软件流量账本"
#ifndef MyAppVersion
#define MyAppVersion "1.0.0"
#endif
#define MyAppExeName "ClashTrafficMonitor.exe"

[Setup]
AppId={{8D166711-F59D-43FC-AD41-6E48887DC552}
AppName=Clash 软件流量账本
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher=Clash Traffic Monitor Community Build
AppPublisherURL=https://github.com/zhf883680/clash-traffic-monitor
AppSupportURL=https://github.com/zhf883680/clash-traffic-monitor/issues
DefaultDirName={localappdata}\ClashTrafficMonitor\app
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
MinVersion=10.0
OutputDir=..\dist
OutputBaseFilename=Clash软件流量账本-Setup-v{#MyAppVersion}
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
LicenseFile=..\LICENSE
UninstallDisplayName=Clash 软件流量账本
UninstallDisplayIcon={app}\{#MyAppExeName}
CloseApplications=yes
CloseApplicationsFilter={#MyAppExeName}
RestartApplications=no
SetupLogging=yes

[Tasks]
Name: "desktopicon"; Description: "创建桌面快捷方式"; GroupDescription: "附加选项："; Flags: checkedonce
Name: "autostart"; Description: "登录 Windows 后自动启动流量账本"; GroupDescription: "附加选项："; Flags: checkedonce

[Files]
Source: "..\dist\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; WorkingDir: "{app}"; Comment: "查看 Clash 中每个软件使用的节点和流量"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; WorkingDir: "{app}"; Comment: "查看 Clash 中每个软件使用的节点和流量"; Tasks: desktopicon

[Registry]
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "ClashTrafficMonitor"; ValueData: """{app}\{#MyAppExeName}"""; Tasks: autostart; Flags: uninsdeletevalue

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "启动 {#MyAppName}"; WorkingDir: "{app}"; Flags: postinstall nowait skipifsilent

[Code]
function PrepareToInstall(var NeedsRestart: Boolean): String;
var
  ResultCode: Integer;
  Attempt: Integer;
  AppPath: String;
begin
  Result := '';
  AppPath := ExpandConstant('{app}\{#MyAppExeName}');
  if not FileExists(AppPath) then
    Exit;

  Exec(AppPath, '--shutdown', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  for Attempt := 1 to 50 do
  begin
    if not CheckForMutexes('Local\ClashTrafficMonitor.SevenLedger') then
      Exit;
    Sleep(100);
  end;

  { Compatibility fallback for versions installed before graceful shutdown existed. }
  Exec(ExpandConstant('{sys}\taskkill.exe'), '/IM {#MyAppExeName} /F', '', SW_HIDE,
    ewWaitUntilTerminated, ResultCode);
  Sleep(300);
  if CheckForMutexes('Local\ClashTrafficMonitor.SevenLedger') then
    Result := '无法关闭正在运行的 Clash 软件流量账本。请从托盘退出后重试。';
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if (CurStep = ssPostInstall) and (not WizardIsTaskSelected('autostart')) then
    RegDeleteValue(HKCU, 'Software\Microsoft\Windows\CurrentVersion\Run', 'ClashTrafficMonitor');
end;

function InitializeUninstall(): Boolean;
begin
  Result := MsgBox(
    '卸载只会删除应用程序和快捷方式。历史流量数据库会继续保留在本机。' + #13#10 + #13#10 +
    '是否继续卸载？',
    mbConfirmation, MB_YESNO) = IDYES;
end;
