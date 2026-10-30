# Windows PIN1 dialog

Native WinUI 3 controls, system theme and DPI scaling. The masked input is 184
effective pixels wide with 24-pixel text. Enter confirms; Escape, Cancel and
the window close button cancel. Empty input cannot be submitted.

The dialog opens centered in the work area of the display containing the mouse
cursor at window creation. It moves to that display before sizing for its DPI,
supports negative desktop coordinates, and does not follow subsequent mouse
movement. If the cursor query fails, it uses the window's nearest display.

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
1 means a dialog failure and 3 means the 60-second PIN deadline expired.
The monotonic countdown starts when the window is loaded, updates the remaining
seconds and progress bar, and checks the deadline again on submission. Expiry
closes the dialog without outputting a PIN. A separate reader watchdog limits
startup or a stalled dialog process to 90 seconds.
The input and temporary output byte buffers are cleared on completion; managed
strings have normal managed-runtime lifetime and cannot promise secure erasure.

For visual testing only, run `EidPd.PinPrompt.exe --preview`. Preview is clearly
labelled, does not access a card and never writes entered text to stdout. Use
dummy input only. Real PIN entry remains a manual hardware acceptance test.
`--check` constructs the window without displaying it or reading a PIN; the
build and installer use it to reject incomplete runtime/resource payloads.

The reader sends bounded JSON context over stdin. The filename in
`challenge.displayContext.fileName` is explicitly labelled as application-supplied.
The name and personal code come only from the selected token's public certificate,
which must match the challenge recipient before asking for a PIN. The caller origin
is supplied by the extension's existing trusted-origin gate. Control characters and
bidi formatting are removed from display strings. Identity is never accepted from
web display metadata. `ERR_EIDPD_PIN_TIMEOUT` is returned on timeout; callers must
fail the pending decrypt job and must not submit a content key.
