using System;
using System.Collections.Generic;
using NATS.Client;
using System.Text.Json;
using LightLink.Types;

namespace LightLink
{
    /// <summary>
    /// LightLink C# Client
    /// Provides RPC, Pub/Sub, State Management, and File Transfer capabilities
    /// </summary>
    public class Client : IDisposable
    {
        private string _url;
        private TLSConfig? _tlsConfig;
        private IConnection? _nc;

        /// <summary>
        /// Create a new client
        /// </summary>
        /// <param name="url">NATS server URL (default: nats://localhost:4222)</param>
        /// <param name="tlsConfig">Optional TLS configuration</param>
        public Client(string url = "nats://localhost:4222", TLSConfig? tlsConfig = null)
        {
            _url = url;
            _tlsConfig = tlsConfig;
        }

        /// <summary>
        /// Connect to NATS server
        /// </summary>
        public void Connect()
        {
            var opts = ConnectionFactory.GetDefaultOptions();
            opts.Url = _url;
            opts.Name = "LightLink C# Client";
            opts.MaxReconnect = 10;
            opts.ReconnectWait = 2000;

            // Configure TLS if provided
            if (_tlsConfig != null)
            {
                ConfigureTLS(opts);
            }

            _nc = new ConnectionFactory().CreateConnection(opts);
        }

        /// <summary>
        /// Connect asynchronously
        /// </summary>
        public System.Threading.Tasks.Task ConnectAsync()
        {
            return System.Threading.Tasks.Task.Run(() => Connect());
        }

        /// <summary>
        /// Close connection
        /// </summary>
        public void Close()
        {
            _nc?.Close();
            _nc = null;
        }

        /// <summary>
        /// Check if connected
        /// </summary>
        public bool IsConnected => _nc != null && _nc.State == ConnState.CONNECTED;

        /// <summary>
        /// Dispose
        /// </summary>
        public void Dispose()
        {
            Close();
        }

        /// <summary>
        /// RPC call (synchronous)
        /// </summary>
        /// <param name="service">Service name</param>
        /// <param name="method">Method name</param>
        /// <param name="args">Arguments dictionary</param>
        /// <param name="timeoutMs">Timeout in milliseconds (default: 5000)</param>
        /// <returns>Result dictionary</returns>
        public Dictionary<string, object> Call(string service, string method,
            Dictionary<string, object> args, int timeoutMs = 5000)
        {
            if (_nc == null)
                throw new InvalidOperationException("Not connected. Call Connect() first.");

            string subject = $"$SRV.{service}.{method}";

            var request = new RPCRequest
            {
                Id = Guid.NewGuid().ToString(),
                Method = method,
                Args = args
            };

            string requestJson = JsonSerializer.Serialize(request);
            byte[] requestData = System.Text.Encoding.UTF8.GetBytes(requestJson);

            try
            {
                Msg msg = _nc.Request(subject, requestData, timeoutMs);
                string responseJson = System.Text.Encoding.UTF8.GetString(msg.Data);

                var response = JsonSerializer.Deserialize<RPCResponse>(responseJson);
                if (response == null || !response.Success)
                {
                    throw new Exception(response?.Error ?? "RPC call failed");
                }

                return response.Result ?? new Dictionary<string, object>();
            }
            catch (NATS.Client.NATSTimeoutException)
            {
                throw new TimeoutException($"RPC call to {service}.{method} timed out");
            }
        }

        /// <summary>
        /// RPC call (asynchronous)
        /// </summary>
        public async System.Threading.Tasks.Task<Dictionary<string, object>> CallAsync(string service, string method,
            Dictionary<string, object> args, int timeoutMs = 5000)
        {
            return await System.Threading.Tasks.Task.Run(() => Call(service, method, args, timeoutMs));
        }

        /// <summary>
        /// Publish message
        /// </summary>
        public void Publish(string subject, Dictionary<string, object> data)
        {
            if (_nc == null)
                throw new InvalidOperationException("Not connected. Call Connect() first.");

            string json = JsonSerializer.Serialize(data);
            byte[] msgData = System.Text.Encoding.UTF8.GetBytes(json);
            _nc.Publish(subject, msgData);
        }

        /// <summary>
        /// Publish message asynchronously
        /// </summary>
        public async System.Threading.Tasks.Task PublishAsync(string subject, Dictionary<string, object> data)
        {
            await System.Threading.Tasks.Task.Run(() => Publish(subject, data));
        }

        /// <summary>
        /// Subscribe to messages
        /// </summary>
        public ISubscription Subscribe(string subject, Action<Dictionary<string, object>> handler)
        {
            if (_nc == null)
                throw new InvalidOperationException("Not connected. Call Connect() first.");

            return _nc.SubscribeAsync(subject, (sender, args) =>
            {
                try
                {
                    var msg = args.Message;
                    string json = System.Text.Encoding.UTF8.GetString(msg.Data);
                    var data = JsonSerializer.Deserialize<Dictionary<string, object>>(json);
                    if (data != null)
                    {
                        handler(data);
                    }
                }
                catch (Exception)
                {
                    // Ignore JSON deserialization errors
                }
            });
        }

        private void ConfigureTLS(Options opts)
        {
            if (_tlsConfig == null) return;

            // Use PFX certificate if available
            if (!string.IsNullOrEmpty(_tlsConfig.PfxFile) &&
                System.IO.File.Exists(_tlsConfig.PfxFile))
            {
                var cert = new System.Security.Cryptography.X509Certificates.X509Certificate2(
                    _tlsConfig.PfxFile,
                    _tlsConfig.PfxPassword);

                opts.AddCertificate(cert);
            }
            // Fall back to cert/key files
            else if (!string.IsNullOrEmpty(_tlsConfig.CertFile) &&
                     System.IO.File.Exists(_tlsConfig.CertFile))
            {
                var cert = new System.Security.Cryptography.X509Certificates.X509Certificate2(
                    _tlsConfig.CertFile);

                opts.AddCertificate(cert);
            }

            // Configure SSL/TLS
            opts.Secure = true;

            // Skip server name verification for self-signed certs
            if (_tlsConfig.InsecureSkipVerify)
            {
                opts.TLSRemoteCertificationValidationCallback =
                    (sender, certificate, chain, sslPolicyErrors) => true;
            }
        }
    }
}
