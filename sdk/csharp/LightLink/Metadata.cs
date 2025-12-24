using System;
using System.Collections.Generic;

namespace LightLink.Metadata
{
    public class ParameterMetadata
    {
        public string Name { get; set; } = "";
        public string Type { get; set; } = "";  // string, number, boolean, array, object
        public bool Required { get; set; }
        public string Description { get; set; } = "";
        public object? Default { get; set; }
    }

    public class ReturnMetadata
    {
        public string Name { get; set; } = "";
        public string Type { get; set; } = "";
        public string Description { get; set; } = "";
    }

    public class ExampleMetadata
    {
        public Dictionary<string, object> Input { get; set; } = new();
        public Dictionary<string, object> Output { get; set; } = new();
        public string Description { get; set; } = "";
    }

    public class MethodMetadata
    {
        public string Name { get; set; } = "";
        public string Description { get; set; } = "";
        public List<ParameterMetadata> Params { get; set; } = new();
        public List<ReturnMetadata> Returns { get; set; } = new();
        public ExampleMetadata? Example { get; set; }
        public List<string> Tags { get; set; } = new();
        public bool Deprecated { get; set; }
    }

    public class ServiceMetadata
    {
        public string Name { get; set; } = "";
        public string Version { get; set; } = "";
        public string Description { get; set; } = "";
        public string Author { get; set; } = "";
        public List<string> Tags { get; set; } = new();
        public List<MethodMetadata> Methods { get; set; } = new();
        public DateTime RegisteredAt { get; set; }
        public DateTime LastSeen { get; set; }
    }
}
