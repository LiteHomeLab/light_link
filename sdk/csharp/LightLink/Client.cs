using System;
using NATS.Client;

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
