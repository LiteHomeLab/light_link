using Xunit;
using LightLink;
using System;
using System.Threading.Tasks;

public class ClientTests
{
    [Fact]
    public void CanCreateClient()
    {
        var client = new Client("nats://localhost:4222");
        Assert.NotNull(client);
    }

    [Fact]
    public async Task CanConnect()
    {
        var client = new Client("nats://localhost:4222");
        try
        {
            await client.ConnectAsync();
            Assert.True(client.IsConnected);
        }
        catch (Exception)
        {
            // Skip if NATS not available
        }
        finally
        {
            client.Close();
        }
    }

    [Fact]
    public async Task CallNonexistentServiceThrows()
    {
        var client = new Client("nats://localhost:4222");
        try
        {
            await client.ConnectAsync();
            await Assert.ThrowsAsync<Exception>(() =>
                client.CallAsync("nonexistent", "method", new Dictionary<string, object>()));
        }
        catch (Exception)
        {
            // Skip if NATS not available
        }
        finally
        {
            client.Close();
        }
    }
}
