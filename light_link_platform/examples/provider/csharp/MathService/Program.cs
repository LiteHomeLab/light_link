using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using LightLink;
using LightLink.Metadata;

namespace MathService
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== C# Metadata Registration Demo ===");

            // Discover client certificates
            Console.WriteLine("\n[1/5] Discovering TLS certificates...");
            var tlsConfig = CertDiscovery.GetAutoTLSConfig();
            Console.WriteLine($"Certificates found:");
            Console.WriteLine($"  CA:   {tlsConfig.CaFile}");
            Console.WriteLine($"  Cert: {tlsConfig.CertFile}");
            Console.WriteLine($"  Key:  {tlsConfig.KeyFile}");

            Console.WriteLine("\n[2/5] Creating service...");
            var svc = new Service("math-service-csharp", "nats://172.18.200.47:4222", tlsConfig);

            // Define method metadata for 'add'
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
                    new() { Name = "sum", Type = "number", Description = "The sum of a and b" }
                },
                Example = new ExampleMetadata
                {
                    Input = new Dictionary<string, object> { { "a", 10 }, { "b", 20 } },
                    Output = new Dictionary<string, object> { { "sum", 30 } },
                    Description = "10 + 20 = 30"
                },
                Tags = new List<string> { "math", "basic", "arithmetic" }
            };

            // Define method metadata for 'multiply'
            var multiplyMeta = new MethodMetadata
            {
                Name = "multiply",
                Description = "Multiply two numbers",
                Params = new List<ParameterMetadata>
                {
                    new() { Name = "a", Type = "number", Required = true, Description = "First factor" },
                    new() { Name = "b", Type = "number", Required = true, Description = "Second factor" }
                },
                Returns = new List<ReturnMetadata>
                {
                    new() { Name = "product", Type = "number", Description = "The product of a and b" }
                },
                Example = new ExampleMetadata
                {
                    Input = new Dictionary<string, object> { { "a", 5 }, { "b", 6 } },
                    Output = new Dictionary<string, object> { { "product", 30 } },
                    Description = "5 * 6 = 30"
                },
                Tags = new List<string> { "math", "basic", "arithmetic" }
            };

            // Define method metadata for 'power'
            var powerMeta = new MethodMetadata
            {
                Name = "power",
                Description = "Calculate a to the power of b",
                Params = new List<ParameterMetadata>
                {
                    new() { Name = "base", Type = "number", Required = true, Description = "The base number" },
                    new() { Name = "exp", Type = "number", Required = true, Description = "The exponent" }
                },
                Returns = new List<ReturnMetadata>
                {
                    new() { Name = "result", Type = "number", Description = "base raised to the power of exp" }
                },
                Example = new ExampleMetadata
                {
                    Input = new Dictionary<string, object> { { "base", 2 }, { "exp", 10 } },
                    Output = new Dictionary<string, object> { { "result", 1024 } },
                    Description = "2^10 = 1024"
                },
                Tags = new List<string> { "math", "advanced" }
            };

            // Define method metadata for 'divide'
            var divideMeta = new MethodMetadata
            {
                Name = "divide",
                Description = "Divide two numbers",
                Params = new List<ParameterMetadata>
                {
                    new() { Name = "numerator", Type = "number", Required = true, Description = "The number to be divided" },
                    new() { Name = "denominator", Type = "number", Required = true, Description = "The number to divide by" }
                },
                Returns = new List<ReturnMetadata>
                {
                    new() { Name = "quotient", Type = "number", Description = "The result of division" }
                },
                Example = new ExampleMetadata
                {
                    Input = new Dictionary<string, object> { { "numerator", 100 }, { "denominator", 4 } },
                    Output = new Dictionary<string, object> { { "quotient", 25 } },
                    Description = "100 / 4 = 25"
                },
                Tags = new List<string> { "math", "basic", "arithmetic" }
            };

            // Register methods with metadata
            Console.WriteLine("\n[3/5] Registering methods with metadata...");
            svc.RegisterMethodWithMetadata("add", AddHandler, addMeta);
            Console.WriteLine("  - add: registered");
            svc.RegisterMethodWithMetadata("multiply", MultiplyHandler, multiplyMeta);
            Console.WriteLine("  - multiply: registered");
            svc.RegisterMethodWithMetadata("power", PowerHandler, powerMeta);
            Console.WriteLine("  - power: registered");
            svc.RegisterMethodWithMetadata("divide", DivideHandler, divideMeta);
            Console.WriteLine("  - divide: registered");

            // Start service first (creates NATS connection)
            Console.WriteLine("\n[4/5] Starting service...");
            svc.Start();
            Console.WriteLine("Service started successfully!");

            // Build and register service metadata (requires NATS connection)
            Console.WriteLine("\n[5/5] Registering service metadata...");
            var metadata = svc.BuildCurrentMetadata(
                "v1.0.0",
                "A mathematical operations service providing basic and advanced math functions (C#)",
                "LiteHomeLab",
                new List<string> { "demo", "math", "calculator", "csharp" }
            );
            svc.RegisterMetadata(metadata);
            Console.WriteLine("Service metadata registered to NATS!");
            Console.WriteLine($"  Service: {metadata.Name}");
            Console.WriteLine($"  Version: {metadata.Version}");
            Console.WriteLine($"  Methods: {metadata.Methods.Count}");

            Console.WriteLine("\n=== Service Information ===");
            Console.WriteLine($"Service Name: {svc.Name}");
            Console.WriteLine($"Registered Methods: {metadata.Methods.Count}");

            Console.WriteLine("\n=== C# Math Service Complete ===");
            Console.WriteLine("\nThe service is now running and will send heartbeat every 30 seconds.");
            Console.WriteLine("Press Ctrl+C to stop the service.");

            await Task.Delay(-1);
        }

        static Task<Dictionary<string, object>> AddHandler(Dictionary<string, object> args)
        {
            double a = Convert.ToDouble(args["a"]);
            double b = Convert.ToDouble(args["b"]);
            return Task.FromResult(new Dictionary<string, object> { { "sum", a + b } });
        }

        static Task<Dictionary<string, object>> MultiplyHandler(Dictionary<string, object> args)
        {
            double a = Convert.ToDouble(args["a"]);
            double b = Convert.ToDouble(args["b"]);
            return Task.FromResult(new Dictionary<string, object> { { "product", a * b } });
        }

        static Task<Dictionary<string, object>> PowerHandler(Dictionary<string, object> args)
        {
            double baseValue = Convert.ToDouble(args["base"]);
            double exp = Convert.ToDouble(args["exp"]);
            double result = 1.0;
            for (int i = 0; i < (int)exp; i++)
            {
                result *= baseValue;
            }
            return Task.FromResult(new Dictionary<string, object> { { "result", result } });
        }

        static Task<Dictionary<string, object>> DivideHandler(Dictionary<string, object> args)
        {
            double numerator = Convert.ToDouble(args["numerator"]);
            double denominator = Convert.ToDouble(args["denominator"]);
            if (denominator == 0)
            {
                throw new ArgumentException("division by zero");
            }
            return Task.FromResult(new Dictionary<string, object> { { "quotient", numerator / denominator } });
        }
    }
}
