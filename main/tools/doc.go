// Package tools provides a set of system tools for the agent to use
// such as file operations, command execution, and file sending.
//
// The tools are organized as follows:
//
//   - ReadFile: Read file contents
//   - WriteFile: Write/create files
//   - AppendToFile: Append content to files
//   - DeleteFile: Delete files
//   - ExecuteCommand: Run shell commands
//   - SendFile: Send files to the user
//
// All tools enforce strict security checks including path traversal prevention,
// file size limits, and execution timeouts.
//
// Configuration is loaded from the agent's work directory which is set in the config.
package tools
