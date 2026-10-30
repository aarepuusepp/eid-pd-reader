using Microsoft.UI.Dispatching;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;

namespace EidPd.PinPrompt;

internal static class Program
{
    internal static bool Preview { get; private set; }
    internal static bool Check { get; private set; }

    [STAThread]
    private static int Main(string[] args)
    {
        // Preview never accesses a card or returns entered text to any process.
        Preview = args.Length == 1 && args[0] == "--preview";
        Check = args.Length == 1 && args[0] == "--check";
        if (args.Length != 0 && !Preview && !Check) return 1;
        Environment.ExitCode = 2;
        try
        {
            WinRT.ComWrappersSupport.InitializeComWrappers();
            Application.Start(parameters =>
            {
                SynchronizationContext.SetSynchronizationContext(
                    new DispatcherQueueSynchronizationContext(DispatcherQueue.GetForCurrentThread()));
                _ = new PinApplication();
            });
        }
        catch
        {
            // Do not log exceptions or input from this security-sensitive process.
            return 1;
        }
        return Environment.ExitCode;
    }
}

public sealed partial class PinApplication : Application
{
    private Window? window;

    public PinApplication()
    {
        InitializeComponent();
        UnhandledException += (_, e) =>
        {
            e.Handled = true;
            Environment.ExitCode = 1;
            Exit();
        };
    }

    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        window = new PinWindow();
        if (Program.Check)
        {
            window.Close();
            Environment.ExitCode = 0;
            Exit();
            return;
        }
        window.Activate();
    }
}
