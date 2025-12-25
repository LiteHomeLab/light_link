using System;
using System.Threading.Tasks;
using NATS.Client;
using NATS.Client.JetStream;

namespace PubSubDemo
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== C# Publish/Subscribe Demo ===");

            // Create NATS connection
            var opts = ConnectionFactory.GetDefaultOptions();
            opts.Url = "nats://localhost:4222";
            opts.Name = "C# PubSub Demo";

            using var conn = new ConnectionFactory().CreateConnection(opts);
            Console.WriteLine("Connected to NATS server");

            // Subscribe to messages
            var subscription = conn.SubscribeSync("test.csharp");
            Console.WriteLine("Subscribed to: test.csharp");

            // Publish test messages
            Console.WriteLine("\nPublishing 3 test messages...");
            for (int i = 1; i <= 3; i++)
            {
                string message = $"Message #{i} from C#";
                conn.Publish("test.csharp", System.Text.Encoding.UTF8.GetBytes(message));
                Console.WriteLine($"  Published: {message}");
                await Task.Delay(500);
            }

            // Receive messages
            Console.WriteLine("\nWaiting for messages (press Ctrl+C to exit)...");
            for (int i = 0; i < 3; i++)
            {
                var msg = subscription.NextMessage(5000);
                if (msg != null)
                {
                    string received = System.Text.Encoding.UTF8.GetString(msg.Data);
                    Console.WriteLine($"  Received: {received}");
                }
            }

            Console.WriteLine("\nDemo complete. Press any key to exit...");
            Console.ReadKey();
        }
    }
}
