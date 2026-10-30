using Microsoft.UI.Dispatching;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using System.Text;
using System.Text.Json;

namespace EidPd.PinPrompt;

internal static class Program
{
    internal static bool Preview { get; private set; }
    internal static bool Check { get; private set; }
    internal static PromptContext Context { get; private set; } = new();

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
            if (Preview)
                Context = new() { FileName = "Näidisdokument.cdoc", Name = "Mari Näidis", PersonalCode = "00000000000", Origin = "https://digisign.example" };
            else if (!Check)
            {
                Console.InputEncoding = Encoding.UTF8;
                // Metadata travels through the reader's inherited stdin pipe,
                // never command-line arguments, files or browser-owned identity.
                var buffer = new char[8193];
                var count = Console.In.ReadBlock(buffer, 0, buffer.Length);
                if (count == buffer.Length) return 1;
                Context = JsonSerializer.Deserialize<PromptContext>(buffer.AsSpan(0, count),
                    new JsonSerializerOptions { PropertyNameCaseInsensitive = true }) ?? new();
                Array.Clear(buffer);
            }
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

internal sealed class PromptContext
{
    public string FileName { get; set; } = "";
    public string Name { get; set; } = "";
    public string PersonalCode { get; set; } = "";
    public string Origin { get; set; } = "";
}

public sealed partial class PinApplication : Application
{
    private Window? window;

    public PinApplication()
    {
        InitializeComponent();
        UnhandledException += (_, e) =>
        {
            if (Program.Check) Console.Error.WriteLine(e.Exception.Message);
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
