// Imports raw event sequences from the protocol's example fixtures.
// Lets the UI run against canonical event traces without a backend.
import textOnly from "@protocols/examples/text-only-stream.json";
import toolCallSuccess from "@protocols/examples/tool-call-success.json";
import toolCallFailure from "@protocols/examples/tool-call-failure.json";
import multiStep from "@protocols/examples/multi-step-investigation.json";

export const fixtures = {
  textOnly,
  toolCallSuccess,
  toolCallFailure,
  multiStep,
} as const;
