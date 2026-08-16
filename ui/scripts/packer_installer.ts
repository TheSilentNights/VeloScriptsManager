import installerCreator from 'electron-winstaller'

installerCreator.createWindowsInstaller({
    appDirectory: 'dist/electron',
    outputDirectory: 'output',
    exe:'dglab-desktop.exe',
})