using System;
using System.Collections.Generic;
using LightLink.Metadata;

namespace LightLink
{
    public class ValidationError
    {
        public string ParameterName { get; set; } = "";
        public string ExpectedType { get; set; } = "";
        public string ActualType { get; set; } = "";
        public object? ActualValue { get; set; }
        public string Message { get; set; } = "";
    }

    public class ValidationResult
    {
        public bool IsValid { get; set; } = true;
        public string ErrorMessage { get; set; } = "";
        public List<ValidationError> Errors { get; set; } = new();
    }

    public class Validator
    {
        private readonly MethodMetadata _methodMeta;

        public Validator(MethodMetadata methodMeta)
        {
            _methodMeta = methodMeta;
        }

        public ValidationResult Validate(Dictionary<string, object> args)
        {
            var result = new ValidationResult();

            foreach (var paramMeta in _methodMeta.Params)
            {
                // 检查必需参数
                if (paramMeta.Required && !args.ContainsKey(paramMeta.Name))
                {
                    result.IsValid = false;
                    result.Errors.Add(new ValidationError
                    {
                        ParameterName = paramMeta.Name,
                        ExpectedType = paramMeta.Type,
                        ActualType = "missing",
                        Message = $"Required parameter '{paramMeta.Name}' is missing"
                    });
                    continue;
                }

                if (!args.ContainsKey(paramMeta.Name))
                    continue;

                // 检查类型
                var value = args[paramMeta.Name];
                string actualType = InferType(value);

                if (!IsTypeCompatible(paramMeta.Type, actualType))
                {
                    result.IsValid = false;
                    result.Errors.Add(new ValidationError
                    {
                        ParameterName = paramMeta.Name,
                        ExpectedType = paramMeta.Type,
                        ActualType = actualType,
                        ActualValue = value,
                        Message = $"Parameter '{paramMeta.Name}': expected type {paramMeta.Type}, got {actualType}"
                    });
                }
            }

            if (result.Errors.Count > 0)
            {
                result.ErrorMessage = $"Validation failed with {result.Errors.Count} error(s)";
            }

            return result;
        }

        private static string InferType(object value)
        {
            if (value == null) return "null";
            Type type = value.GetType();

            if (type == typeof(bool)) return "boolean";
            if (type == typeof(int) || type == typeof(long) ||
                type == typeof(double) || type == typeof(float)) return "number";
            if (type == typeof(string)) return "string";
            return "object";
        }

        private static bool IsTypeCompatible(string expected, string actual)
        {
            return expected == actual;
        }
    }
}
