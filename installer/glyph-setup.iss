; GlyphLang Inno Setup Script
; Creates a Windows installer for GlyphLang

#define MyAppName "GlyphLang"
#define MyAppVersion "0.1.7"
#define MyAppPublisher "GlyphLang"
#define MyAppURL "https://github.com/glyphlang/glyph"
#define MyAppExeName "glyph.exe"

[Setup]
; NOTE: The value of AppId uniquely identifies this application.
AppId={{G1YPH-LANG-0001-0001-000000000001}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}/issues
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
; Output settings
OutputDir=..\dist
OutputBaseFilename=glyph-{#MyAppVersion}-windows-setup
; Compression
Compression=lzma2
SolidCompression=yes
; Installer appearance
WizardStyle=modern
; Privileges - allow per-user install without admin
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
; Minimum Windows version (Windows 10)
MinVersion=10.0
; Uninstall settings
UninstallDisplayIcon={app}\{#MyAppExeName}
UninstallDisplayName={#MyAppName}
; Architecture
ArchitecturesInstallIn64BitMode=x64
ArchitecturesAllowed=x64

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "addtopath"; Description: "Add to PATH (recommended)"; GroupDescription: "Additional options:"; Flags: checkedonce
Name: "desktopicon"; Description: "Create a desktop shortcut"; GroupDescription: "Additional options:"; Flags: unchecked

[Files]
; Main executable
Source: "..\dist\glyph-windows-amd64.exe"; DestDir: "{app}"; DestName: "{#MyAppExeName}"; Flags: ignoreversion

[Icons]
; Start Menu shortcuts
Name: "{group}\GlyphLang Command Prompt"; Filename: "{cmd}"; Parameters: "/k ""{app}\{#MyAppExeName}"" --help"; WorkingDir: "{userdocs}"; Comment: "Open command prompt with GlyphLang"
Name: "{group}\GlyphLang Documentation"; Filename: "{#MyAppURL}"; Comment: "View GlyphLang documentation online"
Name: "{group}\Uninstall GlyphLang"; Filename: "{uninstallexe}"; Comment: "Uninstall GlyphLang"
; Desktop shortcut (optional)
Name: "{userdesktop}\GlyphLang"; Filename: "{cmd}"; Parameters: "/k ""{app}\{#MyAppExeName}"" --help"; WorkingDir: "{userdocs}"; Tasks: desktopicon; Comment: "Open command prompt with GlyphLang"

[Registry]
; File association for .glyph files (optional, user-level)
Root: HKCU; Subkey: "Software\Classes\.glyph"; ValueType: string; ValueName: ""; ValueData: "GlyphLang.Source"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\GlyphLang.Source"; ValueType: string; ValueName: ""; ValueData: "GlyphLang Source File"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Classes\GlyphLang.Source\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName},0"
Root: HKCU; Subkey: "Software\Classes\GlyphLang.Source\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#MyAppExeName}"" ""%1"""

[Code]
// Pascal Script for PATH manipulation

const
  EnvironmentKey = 'Environment';

// Check if a directory is in PATH
function IsDirInPath(Dir: string; Path: string): Boolean;
begin
  Result := Pos(';' + Uppercase(Dir) + ';', ';' + Uppercase(Path) + ';') > 0;
end;

// Add directory to PATH
procedure AddToPath(Dir: string);
var
  Path: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', Path) then
    Path := '';

  if not IsDirInPath(Dir, Path) then
  begin
    if Path <> '' then
      Path := Path + ';';
    Path := Path + Dir;
    RegWriteStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', Path);
  end;
end;

// Remove directory from PATH
procedure RemoveFromPath(Dir: string);
var
  Path, NewPath: string;
  P: Integer;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', Path) then
    Exit;

  NewPath := '';
  while Path <> '' do
  begin
    P := Pos(';', Path);
    if P = 0 then
      P := Length(Path) + 1;

    if Uppercase(Copy(Path, 1, P - 1)) <> Uppercase(Dir) then
    begin
      if NewPath <> '' then
        NewPath := NewPath + ';';
      NewPath := NewPath + Copy(Path, 1, P - 1);
    end;

    Delete(Path, 1, P);
  end;

  RegWriteStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', NewPath);
end;

// Called after installation
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    if IsTaskSelected('addtopath') then
    begin
      AddToPath(ExpandConstant('{app}'));
    end;
  end;
end;

// Called during uninstallation
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    RemoveFromPath(ExpandConstant('{app}'));
  end;
end;

[Messages]
WelcomeLabel2=This will install [name/ver] on your computer.%n%nGlyphLang is an AI-first backend language designed for LLM code generation. Symbol-based syntax, sub-microsecond compilation, built-in security.%n%nIt is recommended that you close all other applications before continuing.

[Run]
; Show help after installation
Filename: "{cmd}"; Parameters: "/k echo GlyphLang installed successfully! && echo. && ""{app}\{#MyAppExeName}"" --version && echo. && echo Type 'glyph --help' to get started. && echo Press any key to close... && pause >nul"; Description: "View GlyphLang version"; Flags: nowait postinstall skipifsilent unchecked
