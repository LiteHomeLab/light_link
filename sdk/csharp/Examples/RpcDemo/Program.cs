using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using LightLink;
using NATS.Client;
using System.Security.Cryptography.X509Certificates;

namespace RpcDemo
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== C# RPC Demo (with TLS) ===");

            // Create TLS options
            var opts = ConnectionFactory.GetDefaultOptions();
            opts.Url = "nats://localhost:4222";
            
            // For NATS.Client with TLS, we need to set up the certificate
            // This requires a newer version or manual certificate handling
            opts.Secure = true;

            // Create service with TLS options
            var svc = new Service("csharp-math-service", "nats://localhost:4222", opts);

            svc.RegisterRPC("add", async (a) =>
            {
                double x = Convert.ToDouble(a["a"]);
                double y = Convert.ToDouble(a["b"]);
                return new Dictionary<string, object> { { "sum", x + y } };
            });

            svc.Start();
            Console.WriteLine("C# Service started. Press Ctrl+C to stop...");

            await Task.Delay(-1);
        }
    }
}
