/*
Package stream provides consumers for RC stream

# Methods
  - Listen establishes a single connection.
  - Run maintains a persistent connection, automatically reconnecting when needed.

# Error handling

If the stream returns errors Listen does not maintains the connection, while Run does.
*/
package stream
