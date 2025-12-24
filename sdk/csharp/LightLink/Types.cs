using System.Collections.Generic;

namespace LightLink.Types
{
    /// <summary>
    /// RPC 请求
    /// </summary>
    public class RPCRequest
    {
        public string Id { get; set; } = "";
        public string Method { get; set; } = "";
        public Dictionary<string, object> Args { get; set; } = new();
    }

    /// <summary>
    /// RPC 响应
    /// </summary>
    public class RPCResponse
    {
        public string Id { get; set; } = "";
        public bool Success { get; set; }
        public Dictionary<string, object>? Result { get; set; }
        public string? Error { get; set; }
    }
}
