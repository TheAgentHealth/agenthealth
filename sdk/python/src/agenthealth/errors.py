"""Tool failures are exceptions; health failures remain results."""


class AgentHealthError(Exception):
    """Base class for SDK/tool errors. Messages never include CLI diagnostics."""


class ConfigurationError(AgentHealthError):
    """Unreadable or invalid SDK configuration."""


class BinaryNotFoundError(AgentHealthError):
    """The requested CLI executable could not be started."""


class CompatibilityError(AgentHealthError):
    """The CLI software version does not match the SDK."""


class ExecutionError(AgentHealthError):
    """The CLI failed to execute or returned exit code 6."""


class CheckTimeoutError(AgentHealthError):
    """The SDK deadline expired; the CLI was stopped."""


class OutputError(AgentHealthError):
    """The CLI returned oversized or invalid machine output."""
