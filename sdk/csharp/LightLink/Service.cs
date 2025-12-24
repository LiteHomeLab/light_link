using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
using NATS.Client;
using System.Text.Json;
using LightLink.Types;
using LightLink.Metadata;

namespace LightLink
{
    public delegate Task<Dictionary<string, object>> RPCHandler(Dictionary<string, object> args);

    public class Service : IDisposable
    {
        private readonly string _name;
        private readonly string _natsURL;
        private IConnection _nc;
        private readonly Dictionary<string, RPCHandler> _rpcHandlers;
        private readonly ReaderWriterLockSlim _rpcLock;
        private readonly Dictionary<string, MethodMetadata> _methodMetadata;
        private readonly ReaderWriterLockSlim _metaLock;
        private Timer? _heartbeatTimer;
        private bool _running;

        private const int HeartbeatIntervalMs = 30000;

        public Service(string name, string natsURL)
        {
            _name = name;
            _natsURL = natsURL;
            _rpcHandlers = new Dictionary<string, RPCHandler>();
            _methodMetadata = new Dictionary<string, MethodMetadata>();
            _rpcLock = new ReaderWriterLockSlim();
            _metaLock = new ReaderWriterLockSlim();
        }

        public string Name => _name;

        public void RegisterRPC(string method, RPCHandler handler)
        {
            _rpcLock.EnterWriteLock();
            try { _rpcHandlers[method] = handler; }
            finally { _rpcLock.ExitWriteLock(); }
        }

        public void RegisterMethodWithMetadata(string method, RPCHandler handler, MethodMetadata metadata)
        {
            _metaLock.EnterWriteLock();
            try { _methodMetadata[method] = metadata; }
            finally { _metaLock.ExitWriteLock(); }
            RegisterRPC(method, handler);
        }

        public bool HasRPC(string method)
        {
            _rpcLock.EnterReadLock();
            try { return _rpcHandlers.ContainsKey(method); }
            finally { _rpcLock.ExitReadLock(); }
        }

        public void Start()
        {
            if (_running) throw new InvalidOperationException("Service already running");

            var opts = ConnectionFactory.GetDefaultOptions();
            opts.Url = _natsURL;
            opts.Name = $"LightLink Service: {_name}";

            _nc = new ConnectionFactory().CreateConnection(opts);

            string subject = $"$SRV.{_name}.>";
            _nc.SubscribeAsync(subject, HandleRPC);

            _heartbeatTimer = new Timer(_ => SendHeartbeat(), null, 0, HeartbeatIntervalMs);

            _running = true;
        }

        private void HandleRPC(object sender, MsgHandlerEventArgs e)
        {
            var msg = e.Message;
            try
            {
                string json = System.Text.Encoding.UTF8.GetString(msg.Data);
                var request = JsonSerializer.Deserialize<RPCRequest>(json);

                _rpcLock.EnterReadLock();
                if (!_rpcHandlers.TryGetValue(request.Method, out var handler))
                {
                    _rpcLock.ExitReadLock();
                    SendError(msg, request.Id, $"Method not found: {request.Method}");
                    return;
                }
                _rpcLock.ExitReadLock();

                var result = handler(request.Args).Result;

                var response = new RPCResponse
                {
                    Id = request.Id,
                    Success = true,
                    Result = result
                };

                string respJson = JsonSerializer.Serialize(response);
                byte[] data = System.Text.Encoding.UTF8.GetBytes(respJson);
                msg.Respond(data);
            }
            catch (Exception ex)
            {
                SendError(msg, "", ex.Message);
            }
        }

        private void SendError(Msg msg, string requestId, string error)
        {
            var response = new RPCResponse
            {
                Id = requestId,
                Success = false,
                Error = error
            };
            string json = JsonSerializer.Serialize(response);
            msg.Respond(System.Text.Encoding.UTF8.GetBytes(json));
        }

        private void SendHeartbeat()
        {
            if (!_running || _nc == null) return;

            var heartbeat = new
            {
                service = _name,
                version = "1.0.0",
                timestamp = DateTime.UtcNow
            };

            string json = JsonSerializer.Serialize(heartbeat);
            _nc.Publish($"$LL.heartbeat.{_name}", System.Text.Encoding.UTF8.GetBytes(json));
        }

        public void Stop()
        {
            if (!_running) return;
            _running = false;
            _heartbeatTimer?.Dispose();
            _nc?.Close();
            _nc = null;
        }

        public void Dispose()
        {
            Stop();
            _rpcLock?.Dispose();
            _metaLock?.Dispose();
        }
    }
}
