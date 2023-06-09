// Code generated by jtd-codegen for C# + System.Text.Json v0.2.1

using System.Text.Json.Serialization;

namespace Senzing
{
    /// <summary>
    /// A detail published by the message generator.
    /// </summary>
    public class Detail
    {
        /// <summary>
        /// The unique identifier of the detail.
        /// </summary>
        [JsonPropertyName("key")]
        public string Key { get; set; }

        /// <summary>
        /// The order in which the detail was given to the message generator.
        /// </summary>
        [JsonPropertyName("position")]
        public int Position { get; set; }

        /// <summary>
        /// Datatype of the value.
        /// </summary>
        [JsonPropertyName("type")]
        public string Type_ { get; set; }

        /// <summary>
        /// The value of the detail in string form.
        /// </summary>
        [JsonPropertyName("value")]
        public string Value { get; set; }

        /// <summary>
        /// The value of the detail if it differs from string form.
        /// </summary>
        [JsonPropertyName("valueRaw")]
        public object ValueRaw { get; set; }
    }
}
