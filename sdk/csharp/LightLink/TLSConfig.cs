using System;
using System.IO;

namespace LightLink
{
    /// <summary>
    /// TLS certificate configuration
    /// </summary>
    public class TLSConfig
    {
        public string CaFile { get; set; } = "";
        public string CertFile { get; set; } = "";
        public string KeyFile { get; set; } = "";
        public string ServerName { get; set; } = "nats-server";

        /// <summary>
        /// PFX certificate file (contains both cert and key).
        /// NATS.Client requires PFX format for TLS connections.
        /// If provided, takes precedence over CertFile/KeyFile.
        /// </summary>
        public string PfxFile { get; set; } = "";

        /// <summary>
        /// Password for PFX file.
        /// </summary>
        public string PfxPassword { get; set; } = "";

        /// <summary>
        /// Skip server certificate name verification for self-signed certificates.
        /// The connection is still encrypted with TLS, and CA chain is verified.
        /// Default: true for compatibility with self-signed certs using legacy CN.
        /// </summary>
        public bool InsecureSkipVerify { get; set; } = true;
    }

    /// <summary>
    /// Certificate discovery result
    /// </summary>
    public class CertDiscoveryResult
    {
        public string CaFile { get; set; } = "";
        public string CertFile { get; set; } = "";
        public string KeyFile { get; set; } = "";
        public string PfxFile { get; set; } = "";
        public string PfxPassword { get; set; } = "lightlink";
        public string ServerName { get; set; } = "nats-server";
        public bool Found { get; set; }
    }

    /// <summary>
    /// Certificate auto-discovery utility
    /// </summary>
    public static class CertDiscovery
    {
        private const string DefaultClientCertDir = "./client";
        private const string DefaultServerCertDir = "./nats-server";
        private const string DefaultServerName = "nats-server";

        /// <summary>
        /// Automatically discover client certificates
        /// </summary>
        public static CertDiscoveryResult DiscoverClientCerts()
        {
            var searchPaths = new[]
            {
                DefaultClientCertDir,
                "../client",
                "../../client",
                "../../../client",
                "../../../../client",
                "../../../../../client"
            };

            foreach (var path in searchPaths)
            {
                var result = CheckCertDirectory(path, "client");
                if (result.Found)
                    return result;
            }

            return new CertDiscoveryResult { Found = false };
        }

        /// <summary>
        /// Automatically discover server certificates
        /// </summary>
        public static CertDiscoveryResult DiscoverServerCerts()
        {
            var searchPaths = new[]
            {
                DefaultServerCertDir,
                "../nats-server",
                "../../nats-server",
                "../../../nats-server",
                "../../../../nats-server",
                "../../../../../nats-server"
            };

            foreach (var path in searchPaths)
            {
                var result = CheckCertDirectory(path, "server");
                if (result.Found)
                    return result;
            }

            return new CertDiscoveryResult { Found = false };
        }

        /// <summary>
        /// Check if certificate files exist in directory
        /// </summary>
        private static CertDiscoveryResult CheckCertDirectory(string dir, string certType)
        {
            var certFile = Path.Combine(dir, certType == "client" ? "client.crt" : "server.crt");
            var keyFile = Path.Combine(dir, certType == "client" ? "client.key" : "server.key");
            var caFile = Path.Combine(dir, "ca.crt");
            var pfxFile = Path.Combine(dir, certType == "client" ? "client.pfx" : "server.pfx");

            // Prefer PFX format for NATS.Client
            if (File.Exists(pfxFile))
            {
                return new CertDiscoveryResult
                {
                    CaFile = caFile,
                    CertFile = certFile,
                    KeyFile = keyFile,
                    PfxFile = pfxFile,
                    ServerName = DefaultServerName,
                    Found = true
                };
            }

            // Fallback to separate cert/key files
            if (File.Exists(caFile) && File.Exists(certFile) && File.Exists(keyFile))
            {
                return new CertDiscoveryResult
                {
                    CaFile = caFile,
                    CertFile = certFile,
                    KeyFile = keyFile,
                    ServerName = DefaultServerName,
                    Found = true
                };
            }

            return new CertDiscoveryResult { Found = false };
        }

        /// <summary>
        /// Convert discovery result to TLSConfig
        /// </summary>
        public static TLSConfig ToTLSConfig(CertDiscoveryResult result)
        {
            return new TLSConfig
            {
                CaFile = result.CaFile,
                CertFile = result.CertFile,
                KeyFile = result.KeyFile,
                PfxFile = result.PfxFile,
                PfxPassword = result.PfxPassword,
                ServerName = result.ServerName
            };
        }

        /// <summary>
        /// Get TLS configuration from auto-discovered certificates.
        /// Throws exception if certificates not found.
        /// </summary>
        public static TLSConfig GetAutoTLSConfig()
        {
            var result = DiscoverClientCerts();
            if (!result.Found)
                throw new InvalidOperationException(
                    "Client certificates not found. Please copy the 'client/' folder from generated certificates to your project.");

            return ToTLSConfig(result);
        }

        /// <summary>
        /// Get server TLS configuration from auto-discovered certificates.
        /// Throws exception if certificates not found.
        /// </summary>
        public static TLSConfig GetAutoServerTLSConfig()
        {
            var result = DiscoverServerCerts();
            if (!result.Found)
                throw new InvalidOperationException(
                    "Server certificates not found. Please copy the 'nats-server/' folder from generated certificates to your project.");

            return ToTLSConfig(result);
        }
    }
}
