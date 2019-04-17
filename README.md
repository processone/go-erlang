# Bert RPC client for Go

Bert stands for Binary ERlang Term. It is a Remote Procedure Call mechanism to support interop between Erlang code\
and other programming languages.

BERT-RPC Go library implements a subset / variation of the BERT-RPC protocol.

Here are the important points to note:
- This version supports BERT-RPC over HTTP. It is not optimal for performance, but can rely on standard HTTP tooling
  features like connection pools, authentication, load balancing, etc.
- This version implements the type I needed in Erlang External Term Format for interop with
  [ejabberd](https://github.com/processone/ejabberd/)

## Installation

## Usage

## TODO

Support various transport for BERT-RPC client:
- HTTP
- TCP/IP
- MQTT
