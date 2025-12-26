using Xunit;
using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace LightLink.Tests
{
    public class ClientTests : IDisposable
    {
        [Fact]
        public void Client_Connect_WithoutTLS_ShouldConnect()
        {
            // Arrange
            var client = new Client("nats://localhost:4222");

            // Act
            client.Connect();

            // Assert
            Assert.True(client.IsConnected);

            // Cleanup
            client.Close();
        }

        [Fact]
        public async Task Client_ConnectAsync_WithoutTLS_ShouldConnect()
        {
            // Arrange
            var client = new Client("nats://localhost:4222");

            // Act
            await client.ConnectAsync();

            // Assert
            Assert.True(client.IsConnected);

            // Cleanup
            client.Close();
        }

        [Fact]
        public async Task Client_Call_ShouldInvokeRemoteMethod()
        {
            // Arrange
            var client = new Client("nats://localhost:4222");
            await client.ConnectAsync();

            // Act
            var result = client.Call("math-service", "add",
                new Dictionary<string, object>
                {
                    { "a", 10 },
                    { "b", 20 }
                });

            // Assert
            Assert.NotNull(result);
            Assert.True(result.ContainsKey("result"));
            Assert.Equal(30, (int)result["result"]);

            // Cleanup
            client.Close();
        }

        [Fact]
        public async Task Client_PublishSubscribe_ShouldReceiveMessage()
        {
            // Arrange
            var client = new Client("nats://localhost:4222");
            await client.ConnectAsync();

            var subject = "test.pubsub";
            var received = false;
            var receivedData = new Dictionary<string, object>();

            // Subscribe first
            using (var sub = client.Subscribe(subject, (data) =>
            {
                received = true;
                receivedData = data;
            }))
            {
                // Wait for subscription to be ready
                await Task.Delay(100);

                // Act
                client.Publish(subject, new Dictionary<string, object>
                {
                    { "message", "Hello, LightLink!" },
                    { "count", 42 }
                });

                // Wait for message
                await Task.Delay(200);

                // Assert
                Assert.True(received);
                Assert.True(receivedData.ContainsKey("message"));
                Assert.Equal("Hello, LightLink!", receivedData["message"]);
                Assert.Equal(42, receivedData["count"]);
            }

            // Cleanup
            client.Close();
        }

        public void Dispose()
        {
            // Cleanup code
        }
    }
}
