using Xunit;
using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace LightLink.Tests
{
    public class ServiceTests : IDisposable
    {
        private Service _svc;

        public ServiceTests()
        {
            _svc = new Service("test-service", "nats://localhost:4222");
        }

        public void Dispose()
        {
            _svc?.Dispose();
        }

        [Fact]
        public void RegisterRPC_AddsHandler()
        {
            _svc.RegisterRPC("test", async (args) => new Dictionary<string, object>());
            Assert.True(_svc.HasRPC("test"));
        }

        [Fact]
        public void Name_ReturnsServiceName()
        {
            Assert.Equal("test-service", _svc.Name);
        }
    }
}
