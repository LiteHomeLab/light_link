using System;
using System.Collections.Generic;
using System.Threading;
using LightLink;

class Program
{
    static void Main(string[] args)
    {
        Console.WriteLine("=== C# SDK Pub/Sub Demo ===");

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

        // Subscribe
        Console.WriteLine("\n[2/2] Testing pub/sub...");
        int receivedCount = 0;

        using (client.Subscribe("test.csharp", (data) =>
        {
            receivedCount++;
            Console.WriteLine($"  Received message #{receivedCount}");
        }))
        {
            // Publish messages
            for (int i = 0; i < 3; i++)
            {
                var msgData = new Dictionary<string, string>
                {
                    { "id", i.ToString() },
                    { "msg", $"Hello from C# #{i}" }
                };

                client.Publish("test.csharp", msgData);
                Console.WriteLine($"  Published message #{i + 1}");
            }

            // Wait for messages
            Thread.Sleep(1000);

            Console.WriteLine($"  Total received: {receivedCount} messages");
        }

        Console.WriteLine("\n=== C# SDK Pub/Sub Demo Complete ===");
    }
}
