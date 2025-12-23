#ifndef LIGHTLINK_CLIENT_HPP
#define LIGHTLINK_CLIENT_HPP

#include <string>
#include <functional>
#include <memory>
#include <map>
#include <vector>
#include <mutex>

namespace lightlink {

// TLS configuration
struct TLSConfig {
    std::string ca_file;
    std::string cert_file;
    std::string key_file;
};

// Forward declarations
class ClientImpl;

// Message handler callback type
using MessageHandler = std::function<void(const std::map<std::string, std::string>&)>;

// RPC result callback type
using RPCCallback = std::function<void(const std::map<std::string, std::string>&, const std::string&)>;

/**
 * LightLink C++ Client
 *
 * Provides RPC, Pub/Sub, State Management, and File Transfer capabilities
 */
class Client {
public:
    /**
     * Create a new client
     * @param url NATS server URL (default: nats://localhost:4222)
     * @param tls_config Optional TLS configuration
     */
    explicit Client(const std::string& url = "nats://localhost:4222",
                    const TLSConfig* tls_config = nullptr);

    ~Client();

    // Disable copy
    Client(const Client&) = delete;
    Client& operator=(const Client&) = delete;

    // Enable move
    Client(Client&&) noexcept;
    Client& operator=(Client&&) noexcept;

    /**
     * Connect to NATS server
     * @return true if successful
     */
    bool connect();

    /**
     * Close connection
     */
    void close();

    /**
     * Check if connected
     */
    bool is_connected() const;

    /**
     * RPC call (synchronous)
     * @param service Service name
     * @param method Method name
     * @param args Arguments map
     * @param timeout_ms Timeout in milliseconds (default: 5000)
     * @return Result map
     */
    std::map<std::string, std::string> call(
        const std::string& service,
        const std::string& method,
        const std::map<std::string, std::string>& args,
        int timeout_ms = 5000);

    /**
     * RPC call (asynchronous)
     * @param service Service name
     * @param method Method name
     * @param args Arguments map
     * @param callback Result callback
     * @param timeout_ms Timeout in milliseconds (default: 5000)
     */
    void call_async(
        const std::string& service,
        const std::string& method,
        const std::map<std::string, std::string>& args,
        RPCCallback callback,
        int timeout_ms = 5000);

    /**
     * Publish message
     * @param subject Subject to publish to
     * @param data Data map
     * @return true if successful
     */
    bool publish(const std::string& subject, const std::map<std::string, std::string>& data);

    /**
     * Subscribe to messages
     * @param subject Subject to subscribe to
     * @param handler Message handler callback
     * @return Subscription ID
     */
    std::string subscribe(const std::string& subject, MessageHandler handler);

    /**
     * Unsubscribe
     * @param subscription_id Subscription ID returned by subscribe()
     */
    void unsubscribe(const std::string& subscription_id);

    /**
     * Set state value
     * @param key State key
     * @param value State value map
     * @return true if successful
     */
    bool set_state(const std::string& key, const std::map<std::string, std::string>& value);

    /**
     * Get state value
     * @param key State key
     * @return State value map
     */
    std::map<std::string, std::string> get_state(const std::string& key);

    /**
     * Watch state changes
     * @param key State key
     * @param handler State change handler
     * @return Watch ID
     */
    std::string watch_state(const std::string& key, MessageHandler handler);

    /**
     * Stop watching state
     * @param watch_id Watch ID
     */
    void unwatch_state(const std::string& watch_id);

    /**
     * Upload file
     * @param file_path Local file path
     * @param remote_name Remote file name
     * @return File ID
     */
    std::string upload_file(const std::string& file_path, const std::string& remote_name);

    /**
     * Download file
     * @param file_id File ID
     * @param local_path Local file path to save
     * @return true if successful
     */
    bool download_file(const std::string& file_id, const std::string& local_path);

private:
    std::unique_ptr<ClientImpl> impl_;
    std::mutex mutex_;
};

} // namespace lightlink

#endif // LIGHTLINK_CLIENT_HPP
