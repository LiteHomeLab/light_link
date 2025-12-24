using System;
using System.Collections.Generic;
using LightLink;

class Program
{
    static void Main(string[] args)
    {
        Console.WriteLine("=== C# SDK RPC Demo ===");

        // Get NATS URL from environment or use default
        string natsUrl = Environment.GetEnvironmentVariable("NATS_URL") ?? "nats://172.18.200.47:4222";
        Console.WriteLine($"NATS URL: {natsUrl}");

        // Create client
        using var client = new Client(natsUrl);

        // Connect
        Console.WriteLine("\n[1/2] Connecting to NATS...");
        try
        {
            client.Connect();
            Console.WriteLine("Connected successfully!");
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Failed to connect: {ex.Message}");
            return;
        }

        // RPC call
        Console.WriteLine("\n[2/2] Testing RPC call...");
        try
        {
            var args = new Dictionary<string, string>
            {
                { "a", "50" },
                { "b", "75" }
            };

            var result = client.Call("demo-service", "add", args);

            if (result.ContainsKey("error"))
            {
                Console.WriteLine($"RPC Error: {result["error"]}");
            }
            else
            {
                Console.WriteLine($"RPC Result: sum={result.GetValueOrDefault("sum", "N/A")}");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"RPC Exception: {ex.Message}");
        }

        Console.WriteLine("\n=== C# SDK RPC Demo Complete ===");
    }
}
