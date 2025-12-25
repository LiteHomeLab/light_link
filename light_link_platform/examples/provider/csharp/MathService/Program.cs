using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using LightLink;
using LightLink.Metadata;
using NATS.Client;

namespace MathService
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== C# Metadata Registration Demo ===");

            // Discover client certificates
            Console.WriteLine("\n[1/4] Discovering TLS certificates...");
            var tlsResult = LightLink.TLSConfig.CertDiscovery.DiscoverClientCerts();
            if (!tlsResult.Found)
            {
                Console.WriteLine("ERROR: Client certificates not found!");
                Console.WriteLine("Please copy the 'client/' folder to your service directory.");
                return;
            }
            Console.WriteLine($"Certificates found:");
            Console.WriteLine($"  CA:   {tlsResult.CaFile}");
            Console.WriteLine($"  Cert: {tlsResult.CertFile}");
            Console.WriteLine($"  Key:  {tlsResult.KeyFile}");

            // Create NATS options with TLS
            var opts = ConnectionFactory.GetDefaultOptions();
            opts.Url = "nats://172.18.200.47:4222";

            // Configure TLS
            var tlsConfig = LightLink.TLSConfig.CertDiscovery.ToTLSConfig(tlsResult);
            opts.SetCertificate(tlsConfig.CaFile, tlsConfig.CertFile, tlsConfig.KeyFile);
            opts.Secure = true;

            var svc = new Service("math-service", "nats://172.18.200.47:4222", opts);

            var addMeta = new MethodMetadata
            {
                Name = "add",
                Description = "Add two numbers together",
                Params = new List<ParameterMetadata>
                {
                    new() { Name = "a", Type = "number", Required = true, Description = "First number" },
                    new() { Name = "b", Type = "number", Required = true, Description = "Second number" }
                },
                Returns = new List<ReturnMetadata>
                {
                    new() { Name = "sum", Type = "number", Description = "The sum" }
                },
                Example = new ExampleMetadata
                {
                    Input = new Dictionary<string, object> { { "a", 10 }, { "b", 20 } },
                    Output = new Dictionary<string, object> { { "sum", 30 } },
                    Description = "10 + 20 = 30"
                }
            };

            svc.RegisterMethodWithMetadata("add", AddHandler, addMeta);
            svc.Start();

            Console.WriteLine("Service with metadata registered. Press Ctrl+C to stop...");
            await Task.Delay(-1);
        }

        static Task<Dictionary<string, object>> AddHandler(Dictionary<string, object> args)
        {
            double a = Convert.ToDouble(args["a"]);
            double b = Convert.ToDouble(args["b"]);
            return Task.FromResult(new Dictionary<string, object> { { "sum", a + b } });
        }
    }
}
