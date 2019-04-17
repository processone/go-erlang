# BERT-RPC library for Go

BERT-RPC library for Go is designed for simple data exchange between Go and Erlang/Elixir applications.

Bert stands for Binary ERlang Term. It is a Remote Procedure Call mechanism to support interop between Erlang code\
and other programming languages.

BERT-RPC Go library implements a subset / variation of the BERT-RPC protocol for Go.

Here are the important points to note:
- This version supports BERT-RPC over HTTP. It is not optimal for performance, but can rely on standard HTTP tooling
  features like connection pools, authentication, load balancing, etc.
- This version implements the type I needed in Erlang External Term Format for interop with
  [ejabberd](https://github.com/processone/ejabberd/)
  
## Why use BERT?

If you want to exchange data with Erlang node, it is handy to use a format that support all the Erlang types, including
atoms. Without having the concept of atoms explicitly in the data exchange protocol, you end up adding wrapper tuples
and conversions on the Erlang side that become very painful.

If you do not need to interop with Erlang, we would recommand using Protobuf or MsgPack.

## Installation

## Usage

## TODO

Support various transport for BERT-RPC client:
- HTTP
- TCP/IP
- MQTT
