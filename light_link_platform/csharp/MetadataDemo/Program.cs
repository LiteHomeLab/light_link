using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using LightLink;
using LightLink.Metadata;

namespace MetadataDemo
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== C# Metadata Registration Demo ===");

            var svc = new Service("math-service", "nats://localhost:4222");

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
