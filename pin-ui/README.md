# Windows PIN1 dialog

Native WinUI 3 controls, system theme and DPI scaling. The masked input is 184
effective pixels wide with 24-pixel text. Enter confirms; Escape, Cancel and
the window close button cancel. Empty input cannot be submitted.

Build with .NET 10 SDK on Windows x64:

```powershell
pwsh -File pin-ui/build.ps1 -OutputDirectory pin-ui/publish
```

The NSIS builder publishes this directory beside the reader as `pin-ui/`.
Both .NET and Windows App SDK are self-contained. No separate runtime install
is needed. Package the entire publish directory, not just the EXE.

The reader resolves the EXE relative to its own installation directory and
captures its inherited stdout pipe. PINs are not passed in arguments, written
to files, or sent to the browser. Exit 0 means confirmed, 2 means cancelled,
1 means a dialog failure. The reader terminates the dialog after 60 seconds.
The input and temporary output byte buffers are cleared on completion; managed
strings have normal managed-runtime lifetime and cannot promise secure erasure.

For visual testing only, run `EidPd.PinPrompt.exe --preview`. Preview is clearly
labelled, does not access a card and never writes entered text to stdout. Use
dummy input only. Real PIN entry remains a manual hardware acceptance test.
`--check` constructs the window without displaying it or reading a PIN; the
build and installer use it to reject incomplete runtime/resource payloads.
