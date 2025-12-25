using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using LightLink;
using LightLink.Metadata;

namespace TextService
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== C# Text Processing Service Demo ===");

            // Discover client certificates
            Console.WriteLine("\n[1/4] Discovering TLS certificates...");
            var tlsConfig = CertDiscovery.GetAutoTLSConfig();
            Console.WriteLine($"Certificates found:");
            Console.WriteLine($"  CA:   {tlsConfig.CaFile}");
            Console.WriteLine($"  Cert: {tlsConfig.CertFile}");
            Console.WriteLine($"  Key:  {tlsConfig.KeyFile}");

            var svc = new Service("csharp-text-service", "nats://172.18.200.47:4222", tlsConfig);

            // Register methods with metadata
            svc.RegisterMethodWithMetadata("reverse", ReverseHandler, new MethodMetadata
            {
                Name = "reverse",
                Description = "Reverse a string",
                Params = new List<ParameterMetadata>
                {
                    new ParameterMetadata { Name = "text", Type = "string", Required = true, Description = "Text to reverse" }
                },
                Returns = new List<ReturnMetadata>
                {
                    new ReturnMetadata { Name = "result", Type = "string", Description = "Reversed text" }
                },
                Example = new ExampleMetadata
                {
                    Input = new Dictionary<string, object> { { "text", "hello" } },
                    Output = new Dictionary<string, object> { { "result", "olleh" } },
                    Description = "Reverse 'hello' to 'olleh'"
                },
                Tags = new List<string> { "string", "transform" }
            });

            svc.RegisterMethodWithMetadata("uppercase", UppercaseHandler, new MethodMetadata
            {
                Name = "uppercase",
                Description = "Convert text to uppercase",
                Params = new List<ParameterMetadata>
                {
                    new ParameterMetadata { Name = "text", Type = "string", Required = true, Description = "Text to convert" }
                },
                Returns = new List<ReturnMetadata>
                {
                    new ReturnMetadata { Name = "result", Type = "string", Description = "Uppercase text" }
                },
                Example = new ExampleMetadata
                {
                    Input = new Dictionary<string, object> { { "text", "hello" } },
                    Output = new Dictionary<string, object> { { "result", "HELLO" } },
                    Description = "Convert 'hello' to 'HELLO'"
                },
                Tags = new List<string> { "string", "transform" }
            });

            svc.RegisterMethodWithMetadata("wordcount", WordCountHandler, new MethodMetadata
            {
                Name = "wordcount",
                Description = "Count words in text",
                Params = new List<ParameterMetadata>
                {
                    new ParameterMetadata { Name = "text", Type = "string", Required = true, Description = "Text to analyze" }
                },
                Returns = new List<ReturnMetadata>
                {
                    new ReturnMetadata { Name = "count", Type = "number", Description = "Number of words" }
                },
                Example = new ExampleMetadata
                {
                    Input = new Dictionary<string, object> { { "text", "hello world" } },
                    Output = new Dictionary<string, object> { { "count", 2 } },
                    Description = "Count 2 words in 'hello world'"
                },
                Tags = new List<string> { "string", "analysis" }
            });

            svc.Start();

            // Build and register metadata
            var metadata = svc.BuildCurrentMetadata(
                "1.0.0",
                "C# Text Processing Service - String manipulation and analysis",
                "LightLink Team",
                new List<string> { "csharp", "text", "string-processing" }
            );
            svc.RegisterMetadata(metadata);

            Console.WriteLine("C# Text Service started and registered. Press Ctrl+C to stop...");
            Console.WriteLine("Service: csharp-text-service");
            Console.WriteLine("Methods: reverse, uppercase, wordcount");

            await Task.Delay(-1);
        }

        static async Task<Dictionary<string, object>> ReverseHandler(Dictionary<string, object> args)
        {
            string text = args["text"].ToString() ?? "";
            char[] arr = text.ToCharArray();
            Array.Reverse(arr);
            return new Dictionary<string, object> { { "result", new string(arr) } };
        }

        static async Task<Dictionary<string, object>> UppercaseHandler(Dictionary<string, object> args)
        {
            string text = args["text"].ToString() ?? "";
            return new Dictionary<string, object> { { "result", text.ToUpper() } };
        }

        static async Task<Dictionary<string, object>> WordCountHandler(Dictionary<string, object> args)
        {
            string text = args["text"].ToString() ?? "";
            string[] words = text.Split(new[] { ' ', '\t', '\n', '\r' }, StringSplitOptions.RemoveEmptyEntries);
            return new Dictionary<string, object> { { "count", words.Length } };
        }
    }
}
