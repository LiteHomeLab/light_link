using System;
using System.Threading.Tasks;
using NATS.Client;
using NATS.Client.JetStream;
using LightLink;

namespace PubSubDemo
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== C# Publish/Subscribe Demo ===");

            // Discover client certificates
            Console.WriteLine("\n[1/3] Discovering TLS certificates...");
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

            // Create NATS connection with TLS
            Console.WriteLine("\n[2/3] Connecting to NATS with TLS...");
            var opts = ConnectionFactory.GetDefaultOptions();
            opts.Url = "nats://172.18.200.47:4222";
            opts.Name = "C# PubSub Demo";

            // Configure TLS
            var tlsConfig = LightLink.TLSConfig.CertDiscovery.ToTLSConfig(tlsResult);
            opts.SetCertificate(tlsConfig.CaFile, tlsConfig.CertFile, tlsConfig.KeyFile);
            opts.Secure = true;

            using var conn = new ConnectionFactory().CreateConnection(opts);
            Console.WriteLine("Connected to NATS server with TLS!");

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
