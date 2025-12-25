using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace LightLink.Types
{
    /// <summary>
    /// RPC 请求
    /// </summary>
    public class RPCRequest
    {
        [JsonPropertyName("id")]
        public string Id { get; set; } = "";

        [JsonPropertyName("method")]
        public string Method { get; set; } = "";

        [JsonPropertyName("args")]
        public Dictionary<string, object> Args { get; set; } = new();
    }

    /// <summary>
    /// RPC 响应
    /// </summary>
    public class RPCResponse
    {
        [JsonPropertyName("id")]
        public string Id { get; set; } = "";

        [JsonPropertyName("success")]
        public bool Success { get; set; }

        [JsonPropertyName("result")]
        public Dictionary<string, object>? Result { get; set; }

        [JsonPropertyName("error")]
        public string? Error { get; set; }
    }
}
