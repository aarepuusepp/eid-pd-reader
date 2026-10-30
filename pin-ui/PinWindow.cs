using Microsoft.UI;
using Microsoft.UI.Windowing;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Automation;
using Microsoft.UI.Xaml.Controls;
using Microsoft.UI.Xaml.Input;
using Microsoft.UI.Xaml.Media;
using System.Runtime.InteropServices;
using System.Text;
using System.Diagnostics;
using Windows.Graphics;
using Windows.System;

namespace EidPd.PinPrompt;

internal sealed class PinWindow : Window
{
    private readonly PasswordBox input;
    private readonly Button confirm;
    private bool submitted;
    private readonly Stopwatch elapsed = new();
    private readonly DispatcherTimer timer = new() { Interval = TimeSpan.FromMilliseconds(200) };
    private readonly ProgressBar remainingBar = new() { Minimum = 0, Maximum = 60, Value = 60, Height = 6, HorizontalAlignment = HorizontalAlignment.Stretch };
    private readonly TextBlock remainingLabel = new() { Text = "PIN1 kinnitamiseks on aega 60 sekundit", FontSize = 14 };
    private bool countdownStarted;

    [DllImport("user32.dll")]
    private static extern uint GetDpiForWindow(nint hwnd);

    [DllImport("user32.dll")]
    [return: MarshalAs(UnmanagedType.Bool)]
    private static extern bool GetCursorPos(out PointInt32 point);

    [StructLayout(LayoutKind.Sequential)]
    private struct NativeRect { public int Left, Top, Right, Bottom; }

    [StructLayout(LayoutKind.Sequential)]
    private struct MonitorInfo
    {
        public uint Size;
        public NativeRect Monitor, Work;
        public uint Flags;
    }

    [DllImport("user32.dll")]
    private static extern nint MonitorFromPoint(PointInt32 point, uint flags);

    [DllImport("user32.dll", EntryPoint = "GetMonitorInfoW")]
    [return: MarshalAs(UnmanagedType.Bool)]
    private static extern bool GetMonitorInfo(nint monitor, ref MonitorInfo info);

    [DllImport("user32.dll")]
    [return: MarshalAs(UnmanagedType.Bool)]
    private static extern bool GetWindowRect(nint hwnd, out NativeRect rect);

    private delegate bool MonitorCallback(nint monitor, nint dc, nint rect, nint data);

    [DllImport("user32.dll")]
    [return: MarshalAs(UnmanagedType.Bool)]
    private static extern bool EnumDisplayMonitors(nint dc, nint clip, MonitorCallback callback, nint data);

    public PinWindow()
    {
        Title = Program.Preview ? "EID-PD · Kujunduse eelvaade" : "EID-PD · PIN1";
        SystemBackdrop = new MicaBackdrop();

        var root = new Grid { Padding = new Thickness(28), RowSpacing = 16 };
        for (var i = 0; i < 6; i++) root.RowDefinitions.Add(new RowDefinition { Height = GridLength.Auto });

        var heading = new StackPanel { Orientation = Orientation.Horizontal, Spacing = 12 };
        heading.Children.Add(new FontIcon { Glyph = "\uE72E", FontSize = 24, VerticalAlignment = VerticalAlignment.Center });
        heading.Children.Add(new TextBlock
        {
            Text = "Sisesta PIN1", FontSize = 24, FontWeight = Microsoft.UI.Text.FontWeights.SemiBold,
            VerticalAlignment = VerticalAlignment.Center
        });
        root.Children.Add(heading);

        var description = new TextBlock
        {
            Text = Program.Preview
                ? "Kujunduse eelvaade. Kasuta ainult näidisnumbreid; kaarti ei kasutata."
                : "Kinnita dekrüpteerimine oma ID-kaardiga.",
            FontSize = 16, TextWrapping = TextWrapping.Wrap
        };
        Grid.SetRow(description, 1);
        root.Children.Add(description);

        var details = new StackPanel { Spacing = 5 };
        details.Children.Add(new TextBlock { Text = "Fail · veebirakenduse esitatud nimi", FontSize = 13 });
        details.Children.Add(new TextBlock
        {
            Text = string.IsNullOrWhiteSpace(Program.Context.FileName) ? "Failinime ei edastatud" : Program.Context.FileName,
            FontSize = 16, TextWrapping = TextWrapping.Wrap, MaxLines = 2, TextTrimming = TextTrimming.CharacterEllipsis
        });
        details.Children.Add(new TextBlock { Text = "Dekrüpteerija · ID-kaardi sertifikaadilt", FontSize = 13, Margin = new Thickness(0, 7, 0, 0) });
        details.Children.Add(new TextBlock
        {
            Text = string.IsNullOrWhiteSpace(Program.Context.Name) ? "Isikuandmeid ei edastatud" : Program.Context.Name,
            FontSize = 16, FontWeight = Microsoft.UI.Text.FontWeights.SemiBold,
            TextTrimming = TextTrimming.CharacterEllipsis
        });
        details.Children.Add(new TextBlock { Text = Program.Context.PersonalCode, FontSize = 16 });
        if (!string.IsNullOrWhiteSpace(Program.Context.Origin))
            details.Children.Add(new TextBlock { Text = Program.Context.Origin, FontSize = 13, TextTrimming = TextTrimming.CharacterEllipsis });
        Grid.SetRow(details, 2);
        root.Children.Add(details);

        var field = new StackPanel { Spacing = 10 };
        input = new PasswordBox
        {
            Width = 184, FontSize = 24, Padding = new Thickness(12, 8, 12, 8),
            HorizontalAlignment = HorizontalAlignment.Left,
            PasswordRevealMode = PasswordRevealMode.Hidden,
            Header = new TextBlock { Text = "ID-kaardi PIN1", FontSize = 14 }, TabIndex = 0
        };
        AutomationProperties.SetName(input, "ID-kaardi PIN1");
        input.KeyDown += (_, e) =>
        {
            if (e.Key == VirtualKey.Enter && input.Password.Length > 0)
            {
                e.Handled = true;
                Submit();
            }
        };
        field.Children.Add(input);
        field.Children.Add(new TextBlock
        {
            Text = "PIN jääb sellesse arvutisse. Seda ei saadeta veebilehele.",
            FontSize = 14, TextWrapping = TextWrapping.Wrap,
            Foreground = (Brush)Application.Current.Resources["TextFillColorSecondaryBrush"]
        });
        Grid.SetRow(field, 3);
        root.Children.Add(field);

        var countdown = new StackPanel { Spacing = 8 };
        countdown.Children.Add(remainingLabel);
        countdown.Children.Add(remainingBar);
        AutomationProperties.SetName(remainingBar, "PIN1 sisestamiseks jäänud aeg");
        Grid.SetRow(countdown, 4);
        root.Children.Add(countdown);

        var actions = new StackPanel { Orientation = Orientation.Horizontal, HorizontalAlignment = HorizontalAlignment.Right, Spacing = 12, Margin = new Thickness(0, 4, 0, 0) };
        var cancel = new Button { Content = "Tühista", MinWidth = 112, Height = 40, FontSize = 16, TabIndex = 1 };
        cancel.Click += (_, _) => Close();
        confirm = new Button
        {
            Content = "Kinnita", MinWidth = 112, Height = 40, FontSize = 16,
            IsEnabled = false, TabIndex = 2,
            Style = (Style)Application.Current.Resources["AccentButtonStyle"]
        };
        confirm.Click += (_, _) => Submit();
        input.PasswordChanged += (_, _) => confirm.IsEnabled = input.Password.Length > 0;
        actions.Children.Add(cancel);
        actions.Children.Add(confirm);
        Grid.SetRow(actions, 5);
        root.Children.Add(actions);
        root.AddHandler(UIElement.KeyDownEvent, new KeyEventHandler((_, e) =>
        {
            if (e.Key == VirtualKey.Escape) { e.Handled = true; Close(); }
        }), true);
        timer.Tick += (_, _) => UpdateCountdown();
        root.Loaded += (_, _) =>
        {
            input.Focus(FocusState.Programmatic);
            if (!countdownStarted && !Program.Check)
            {
                countdownStarted = true;
                elapsed.Start();
                timer.Start();
            }
        };
        Closed += (_, _) => { submitted = true; timer.Stop(); elapsed.Stop(); input.Password = ""; };
        Content = new ScrollViewer
        {
            Content = root,
            VerticalScrollBarVisibility = ScrollBarVisibility.Auto,
            HorizontalScrollBarVisibility = ScrollBarVisibility.Disabled
        };

        var hwnd = WinRT.Interop.WindowNative.GetWindowHandle(this);
        var appWindow = AppWindow.GetFromWindowId(Win32Interop.GetWindowIdFromWindow(hwnd));
        if (appWindow.Presenter is OverlappedPresenter presenter)
        {
            presenter.IsResizable = false;
            presenter.IsMaximizable = false;
            presenter.IsMinimizable = false;
        }
        // Choose once, before showing the window: later mouse movement must not
        // move an active PIN dialog. Fall back if the cursor cannot be queried.
        var point = GetCursorPos(out var cursor) ? cursor : appWindow.Position;
        PlaceOnMonitor(appWindow, hwnd, MonitorFromPoint(point, 2));

        if (Program.Check)
        {
            // Exercise the real Windows geometry on EVERY attached monitor.
            // No window is shown and no card or PIN is accessed in this mode.
            var monitors = new List<nint>();
            if (!EnumDisplayMonitors(0, 0, (monitor, _, _, _) => { monitors.Add(monitor); return true; }, 0))
                throw new InvalidOperationException("Cannot enumerate monitors");
            foreach (var monitor in monitors) PlaceOnMonitor(appWindow, hwnd, monitor);
        }
    }

    private static void PlaceOnMonitor(AppWindow appWindow, nint hwnd, nint monitor)
    {
        var info = new MonitorInfo { Size = (uint)Marshal.SizeOf<MonitorInfo>() };
        if (!GetMonitorInfo(monitor, ref info)) throw new InvalidOperationException("Cannot query monitor work area");
        // MONITORINFO.rcWork is already in absolute virtual-desktop coordinates.
        // Never add a display origin to these coordinates (including negatives).
        var area = info.Work;
        var width = area.Right - area.Left;
        var height = area.Bottom - area.Top;
        if (width <= 0 || height <= 0) throw new InvalidOperationException("Invalid monitor work area");
        // First move the hidden window to its target display so PerMonitorV2
        // applies that display's DPI before calculating the client size.
        appWindow.Move(new PointInt32(area.Left, area.Top));
        var scale = GetDpiForWindow(hwnd) / 96.0;
        appWindow.ResizeClient(new SizeInt32((int)(520 * scale), (int)(560 * scale)));
        appWindow.Resize(new SizeInt32(Math.Min(width, appWindow.Size.Width), Math.Min(height, appWindow.Size.Height)));
        appWindow.Move(new PointInt32(area.Left + (width - appWindow.Size.Width) / 2,
            area.Top + (height - appWindow.Size.Height) / 2));

        if (Program.Check)
        {
            var found = GetWindowRect(hwnd, out var actual);
            Console.WriteLine($"CHECK placement work=({area.Left},{area.Top},{area.Right},{area.Bottom}) window=({actual.Left},{actual.Top},{actual.Right},{actual.Bottom})");
            if (!found || actual.Left < area.Left || actual.Top < area.Top ||
                actual.Right > area.Right || actual.Bottom > area.Bottom ||
                Math.Abs((actual.Left + actual.Right) - (area.Left + area.Right)) > 2 ||
                Math.Abs((actual.Top + actual.Bottom) - (area.Top + area.Bottom)) > 2)
                throw new InvalidOperationException("PIN window is not centered within the monitor work area");
            Console.WriteLine("PASS placement");
        }
    }

    private void Submit()
    {
        // Check the monotonic deadline on submit too, so delayed UI timer ticks
        // cannot accept a PIN after expiry.
        UpdateCountdown();
        if (submitted || input.Password.Length == 0) return;
        submitted = true;
        if (!Program.Preview)
        {
            // Anonymous inherited stdout pipe goes only to the reader. Never use
            // command-line arguments, a file, clipboard, logging, or browser IPC.
            var bytes = Encoding.UTF8.GetBytes(input.Password);
            try
            {
                using var output = Console.OpenStandardOutput();
                output.Write(bytes);
                output.Flush();
            }
            finally { Array.Clear(bytes); }
        }
        input.Password = "";
        Environment.ExitCode = 0;
        Close();
    }

    private void UpdateCountdown()
    {
        if (submitted || !countdownStarted) return;
        var remaining = Math.Max(0, 60 - elapsed.Elapsed.TotalSeconds);
        remainingBar.Value = remaining;
        remainingLabel.Text = $"PIN1 kinnitamiseks on aega {(int)Math.Ceiling(remaining)} sekundit";
        if (remaining > 0) return;
        submitted = true;
        confirm.IsEnabled = false;
        input.Password = "";
        Environment.ExitCode = 3;
        Close();
    }
}
