using Xunit;
using System;
using System.Threading.Tasks;

namespace LightLink.Tests
{
    public class IntegrationTests
    {
        [Fact(Skip = "Requires NATS server")]
        public void CSharpService_RespondsToRPC()
        {
            var svc = new Service("test-service", "nats://localhost:4222");
            svc.RegisterRPC("echo", async (args) =>
                new Dictionary<string, object> { { "echo", args["input"] } });
            svc.Start();

            // Add actual RPC call test here
            svc.Stop();
        }
    }
}
