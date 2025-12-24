using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using LightLink;

namespace RpcDemo
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== C# RPC Demo ===");

            var svc = new Service("math-service", "nats://localhost:4222");

            svc.RegisterRPC("add", async (a) =>
            {
                double x = Convert.ToDouble(a["a"]);
                double y = Convert.ToDouble(a["b"]);
                return new Dictionary<string, object> { { "sum", x + y } };
            });

            svc.Start();
            Console.WriteLine("Service started. Press Ctrl+C to stop...");

            await Task.Delay(-1);
        }
    }
}
