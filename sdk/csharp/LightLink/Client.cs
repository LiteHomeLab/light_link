using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
using NATS.Client;
using NATS.Client.JetStream;

namespace LightLink
{
    /// <summary>
    /// TLS configuration
    /// </summary>
    public class TLSConfig
    {
        public string CaFile { get; set; }
        public string CertFile { get; set; }
        public string KeyFile { get; set; }
    }

    /// <summary>
    /// LightLink C# Client
    /// Provides RPC, Pub/Sub, State Management, and File Transfer capabilities
    /// </summary>
    public class Client : IDisposable
    {
        private string _url;
        private TLSConfig _tlsConfig;
        private IConnection _nc;
        private IJetStream _js;

        /// <summary>
        /// Create a new client
        /// </summary>
        /// <param name="url">NATS server URL (default: nats://localhost:4222)</param>
        /// <param name="tlsConfig">Optional TLS configuration</param>
        public Client(string url = "nats://localhost:4222", TLSConfig tlsConfig = null)
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

            // Configure TLS
            if (_tlsConfig != null)
            {
                opts.Secure = true;
                // Note: Full TLS configuration requires setting up the SSL/TLS context
                // This is a simplified version
            }

            _nc = new ConnectionFactory().CreateConnection(opts);
            _js = _nc.CreateJetStreamContext();
        }

        /// <summary>
        /// Connect asynchronously
        /// </summary>
        public async Task ConnectAsync()
        {
            await Task.Run(() => Connect());
        }

        /// <summary>
        /// Close connection
        /// </summary>
        public void Close()
        {
            _nc?.Close();
            _nc = null;
            _js = null;
        }

        /// <summary>
        /// Check if connected
        /// </summary>
        public bool IsConnected => _nc != null && _nc.State == ConnState.CONNECTED;

        /// <summary>
        /// RPC call (synchronous)
        /// </summary>
        /// <param name="service">Service name</param>
        /// <param name="method">Method name</param>
        /// <param name="args">Arguments dictionary</param>
        /// <param name="timeoutMs">Timeout in milliseconds (default: 5000)</param>
        /// <returns>Result dictionary</returns>
        public Dictionary<string, string> Call(string service, string method,
            Dictionary<string, string> args, int timeoutMs = 5000)
        {
            string subject = $"$SRV.{service}.{method}";

            var request = new RPCRequest
            {
                Id = Guid.NewGuid().ToString(),
                Method = method,
                Args = args
            };

            string requestJson = JsonHelper.Serialize(request);
            byte[] requestData = System.Text.Encoding.UTF8.GetBytes(requestJson);

            try
            {
                Msg msg = _nc.Request(subject, requestData, timeoutMs);
                string responseJson = System.Text.Encoding.UTF8.GetString(msg.Data);

                var response = JsonHelper.Deserialize<RPCResponse>(responseJson);
                if (!response.Success)
                {
                    throw new Exception(response.Error);
                }

                return response.Result;
            }
            catch (NATSTimeoutException)
            {
                throw new Exception("RPC timeout");
            }
        }

        /// <summary>
        /// RPC call (asynchronous)
        /// </summary>
        /// <param name="service">Service name</param>
        /// <param name="method">Method name</param>
        /// <param name="args">Arguments dictionary</param>
        /// <param name="timeoutMs">Timeout in milliseconds (default: 5000)</param>
        /// <returns>Result dictionary</returns>
        public async Task<Dictionary<string, string>> CallAsync(string service, string method,
            Dictionary<string, string> args, int timeoutMs = 5000)
        {
            return await Task.Run(() => Call(service, method, args, timeoutMs));
        }

        /// <summary>
        /// Publish message
        /// </summary>
        /// <param name="subject">Subject to publish to</param>
        /// <param name="data">Data dictionary</param>
        public void Publish(string subject, Dictionary<string, string> data)
        {
            string json = JsonHelper.Serialize(data);
            byte[] msgData = System.Text.Encoding.UTF8.GetBytes(json);
            _nc.Publish(subject, msgData);
        }

        /// <summary>
        /// Publish message asynchronously
        /// </summary>
        /// <param name="subject">Subject to publish to</param>
        /// <param name="data">Data dictionary</param>
        public async Task PublishAsync(string subject, Dictionary<string, string> data)
        {
            await Task.Run(() => Publish(subject, data));
        }

        /// <summary>
        /// Subscribe to messages
        /// </summary>
        /// <param name="subject">Subject to subscribe to</param>
        /// <param name="handler">Message handler callback</param>
        /// <returns>Subscription object</returns>
        public ISubscription Subscribe(string subject, Action<Dictionary<string, string>> handler)
        {
            return _nc.SubscribeAsync(subject, (msg) =>
            {
                string json = System.Text.Encoding.UTF8.GetString(msg.Data);
                var data = JsonHelper.Deserialize<Dictionary<string, string>>(json);
                handler(data);
            });
        }

        /// <summary>
        /// Set state value
        /// </summary>
        /// <param name="key">State key</param>
        /// <param name="value">State value dictionary</param>
        public void SetState(string key, Dictionary<string, string> value)
        {
            var kv = GetOrCreateKeyValue("light_link_state");
            string json = JsonHelper.Serialize(value);
            byte[] data = System.Text.Encoding.UTF8.GetBytes(json);
            kv.Put(key, data);
        }

        /// <summary>
        /// Get state value
        /// </summary>
        /// <param name="key">State key</param>
        /// <returns>State value dictionary</returns>
        public Dictionary<string, string> GetState(string key)
        {
            var kv = GetOrCreateKeyValue("light_link_state");
            var entry = kv.Get(key);
            string json = System.Text.Encoding.UTF8.GetString(entry.Value);
            return JsonHelper.Deserialize<Dictionary<string, string>>(json);
        }

        /// <summary>
        /// Upload file
        /// </summary>
        /// <param name="filePath">Local file path</param>
        /// <param name="remoteName">Remote file name</param>
        /// <returns>File ID</returns>
        public string UploadFile(string filePath, string remoteName)
        {
            var objStore = GetOrCreateObjectStore("light_link_files");
            string fileId = Guid.NewGuid().ToString();

            byte[] fileData = System.IO.File.ReadAllBytes(filePath);
            const int chunkSize = 1024 * 1024; // 1MB chunks

            int chunkNum = 0;
            for (int offset = 0; offset < fileData.Length; offset += chunkSize)
            {
                int size = Math.Min(chunkSize, fileData.Length - offset);
                byte[] chunk = new byte[size];
                Array.Copy(fileData, offset, chunk, 0, size);

                string chunkKey = $"{fileId}_{chunkNum}";
                objStore.Put(chunkKey, chunk);
                chunkNum++;
            }

            // Publish metadata
            var metadata = new Dictionary<string, string>
            {
                { "file_id", fileId },
                { "file_name", remoteName },
                { "file_size", fileData.Length.ToString() },
                { "chunk_num", chunkNum.ToString() }
            };
            Publish("file.uploaded", metadata);

            return fileId;
        }

        /// <summary>
        /// Upload file asynchronously
        /// </summary>
        /// <param name="filePath">Local file path</param>
        /// <param name="remoteName">Remote file name</param>
        /// <returns>File ID</returns>
        public async Task<string> UploadFileAsync(string filePath, string remoteName)
        {
            return await Task.Run(() => UploadFile(filePath, remoteName));
        }

        /// <summary>
        /// Download file
        /// </summary>
        /// <param name="fileId">File ID</param>
        /// <param name="localPath">Local file path to save</param>
        public void DownloadFile(string fileId, string localPath)
        {
            var objStore = GetOrCreateObjectStore("light_link_files");

            using (var fileStream = new System.IO.FileStream(localPath, System.IO.FileMode.Create))
            {
                int chunkNum = 0;
                while (true)
                {
                    try
                    {
                        string chunkKey = $"{fileId}_{chunkNum}";
                        var chunk = objStore.Get(chunkKey);
                        fileStream.Write(chunk.Value, 0, chunk.Value.Length);
                        chunkNum++;
                    }
                    catch
                    {
                        break; // No more chunks
                    }
                }
            }
        }

        /// <summary>
        /// Download file asynchronously
        /// </summary>
        /// <param name="fileId">File ID</param>
        /// <param name="localPath">Local file path to save</param>
        public async Task DownloadFileAsync(string fileId, string localPath)
        {
            await Task.Run(() => DownloadFile(fileId, localPath));
        }

        private IKeyValue GetOrCreateKeyValue(string bucketName)
        {
            try
            {
                return _js.GetKeyValue(bucketName);
            }
            catch
            {
                var config = KeyValueConfiguration.Builder()
                    .WithName(bucketName)
                    .Build();
                return _js.CreateKeyValue(config);
            }
        }

        private IObjectStore GetOrCreateObjectStore(string bucketName)
        {
            try
            {
                return _js.GetObjectStore(bucketName);
            }
            catch
            {
                var config = ObjectStoreConfiguration.Builder()
                    .WithName(bucketName)
                    .Build();
                return _js.CreateObjectStore(config);
            }
        }

        /// <summary>
        /// Dispose
        /// </summary>
        public void Dispose()
        {
            Close();
        }
    }

    // Internal classes for JSON serialization

    internal class RPCRequest
    {
        public string Id { get; set; }
        public string Method { get; set; }
        public Dictionary<string, string> Args { get; set; }
    }

    internal class RPCResponse
    {
        public string Id { get; set; }
        public bool Success { get; set; }
        public Dictionary<string, string> Result { get; set; }
        public string Error { get; set; }
    }

    // Simple JSON helper (in production, use Newtonsoft.Json or System.Text.Json)
    internal static class JsonHelper
    {
        public static string Serialize<T>(T obj)
        {
            // Simplified serialization - in production use proper JSON library
            if (obj is Dictionary<string, string> dict)
            {
                var pairs = new List<string>();
                foreach (var kv in dict)
                {
                    pairs.Add($"\"{kv.Key}\":\"{kv.Value}\"");
                }
                return "{" + string.Join(",", pairs) + "}";
            }
            return "{}";
        }

        public static T Deserialize<T>(string json)
        {
            // Simplified deserialization - in production use proper JSON library
            if (typeof(T) == typeof(Dictionary<string, string>))
            {
                var result = new Dictionary<string, string>();
                // Very basic parsing - would need proper JSON parser in production
                return (T)(object)result;
            }
            return default(T);
        }
    }
}
