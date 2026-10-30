using Microsoft.UI;
using Microsoft.UI.Windowing;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Automation;
using Microsoft.UI.Xaml.Controls;
using Microsoft.UI.Xaml.Input;
using Microsoft.UI.Xaml.Media;
using System.Runtime.InteropServices;
using System.Text;
using Windows.Graphics;
using Windows.System;

namespace EidPd.PinPrompt;

internal sealed class PinWindow : Window
{
    private readonly PasswordBox input;
    private readonly Button confirm;
    private bool submitted;

    [DllImport("user32.dll")]
    private static extern uint GetDpiForWindow(nint hwnd);

    public PinWindow()
    {
        Title = Program.Preview ? "EID-PD · Kujunduse eelvaade" : "EID-PD · PIN1";
        SystemBackdrop = new MicaBackdrop();

        var root = new Grid { Padding = new Thickness(28), RowSpacing = 20 };
        for (var i = 0; i < 4; i++) root.RowDefinitions.Add(new RowDefinition { Height = GridLength.Auto });

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
        Grid.SetRow(field, 2);
        root.Children.Add(field);

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
        Grid.SetRow(actions, 3);
        root.Children.Add(actions);
        root.AddHandler(UIElement.KeyDownEvent, new KeyEventHandler((_, e) =>
        {
            if (e.Key == VirtualKey.Escape) { e.Handled = true; Close(); }
        }), true);
        root.Loaded += (_, _) => input.Focus(FocusState.Programmatic);
        Closed += (_, _) => input.Password = "";
        Content = root;

        var hwnd = WinRT.Interop.WindowNative.GetWindowHandle(this);
        var appWindow = AppWindow.GetFromWindowId(Win32Interop.GetWindowIdFromWindow(hwnd));
        if (appWindow.Presenter is OverlappedPresenter presenter)
        {
            presenter.IsResizable = false;
            presenter.IsMaximizable = false;
            presenter.IsMinimizable = false;
        }
        var scale = GetDpiForWindow(hwnd) / 96.0;
        appWindow.ResizeClient(new SizeInt32((int)(460 * scale), (int)(340 * scale)));
        var display = DisplayArea.GetFromWindowId(appWindow.Id, DisplayAreaFallback.Nearest);
        var area = display.WorkArea;
        appWindow.Move(new PointInt32(area.X + (area.Width - appWindow.Size.Width) / 2,
            area.Y + (area.Height - appWindow.Size.Height) / 2));
    }

    private void Submit()
    {
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
}
