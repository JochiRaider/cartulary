// Authored structural schema components; vocabulary comes from the definitions catalog.
export const instrumentationShapes = {
  "sample_value": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
      "value": {
        "type": [
          "number",
          "null"
        ],
        "minimum": 0
      },
      "availability": {
        "$ref": "#/$defs/availability"
      }
    },
    "required": [
      "value",
      "availability"
    ],
    "allOf": [
      {
        "if": {
          "properties": {
            "availability": {
              "const": "available"
            }
          }
        },
        "then": {
          "properties": {
            "value": {
              "type": "number",
              "minimum": 0
            }
          }
        },
        "else": {
          "properties": {
            "value": {
              "type": "null"
            }
          }
        }
      }
    ]
  },
  "aggregate_value": {
    "type": "object",
    "additionalProperties": false,
    "required": [
      "maximum",
      "observed_delta",
      "observed_interval_ms",
      "unavailable"
    ],
    "properties": {
      "maximum": {
        "type": [
          "number",
          "null"
        ],
        "minimum": 0
      },
      "observed_delta": {
        "type": [
          "number",
          "null"
        ],
        "minimum": 0
      },
      "observed_interval_ms": {
        "type": "number",
        "minimum": 0
      },
      "unavailable": {
        "type": "integer",
        "minimum": 0
      }
    }
  },
  "observer": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
      "cpu_user_us": {
        "type": "number",
        "minimum": 0
      },
      "cpu_system_us": {
        "type": "number",
        "minimum": 0
      },
      "heap_peak_bytes": {
        "type": "number",
        "minimum": 0
      },
      "read_bytes": {
        "type": "number",
        "minimum": 0
      },
      "write_bytes": {
        "type": "number",
        "minimum": 0
      },
      "maximum_sweep_ms": {
        "type": "number",
        "minimum": 0
      },
      "shutdown_ms": {
        "type": "number",
        "minimum": 0
      },
      "gate_cpu": {
        "const": "not_observed"
      }
    },
    "required": [
      "cpu_user_us",
      "cpu_system_us",
      "heap_peak_bytes",
      "read_bytes",
      "write_bytes",
      "maximum_sweep_ms",
      "shutdown_ms",
      "gate_cpu"
    ]
  },
  "local_timing_consistency": {
    "if": {
      "properties": {
        "availability": {
          "const": "available"
        }
      },
      "required": [
        "availability"
      ]
    },
    "then": {
      "properties": {
        "duration_ms": {
          "type": "number"
        },
        "clock_identity": {
          "type": "string"
        },
        "resolution_ms": {
          "type": "number"
        },
        "clock": {
          "not": {
            "const": "unavailable"
          }
        }
      }
    },
    "else": {
      "properties": {
        "duration_ms": {
          "type": "null"
        },
        "clock_identity": {
          "type": "null"
        },
        "resolution_ms": {
          "type": "null"
        },
        "clock": {
          "const": "unavailable"
        }
      }
    }
  }
};
