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

        public void Dispose()
        {
            // Cleanup code
        }
    }
}
