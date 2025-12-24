using Xunit;
using LightLink;
using LightLink.Metadata;

namespace LightLink.Tests
{
    public class ValidatorTests
    {
        [Fact]
        public void Validate_WithValidParams_ReturnsValid()
        {
            var meta = new MethodMetadata
            {
                Params = new List<ParameterMetadata>
                {
                    new() { Name = "a", Type = "number", Required = true }
                }
            };
            var validator = new Validator(meta);
            var args = new Dictionary<string, object> { { "a", 42 } };

            var result = validator.Validate(args);

            Assert.True(result.IsValid);
        }

        [Fact]
        public void Validate_WithMissingRequired_ReturnsInvalid()
        {
            var meta = new MethodMetadata
            {
                Params = new List<ParameterMetadata>
                {
                    new() { Name = "a", Type = "number", Required = true }
                }
            };
            var validator = new Validator(meta);
            var args = new Dictionary<string, object>();

            var result = validator.Validate(args);

            Assert.False(result.IsValid);
            Assert.Single(result.Errors);
            Assert.Equal("missing", result.Errors[0].ActualType);
        }
    }
}
