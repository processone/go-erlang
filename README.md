# Go-Erlang

Go-Erlang is a set of tools for Go <-> Erlang interoperability.

The core of the library is the Erlang External Term Format. It is the internal format data to exchange Erlang terms 
over the network. This binary format is used for example in Erlang distribution protocol.

## Installation

## Usage

## BERT and BERT-RPC library for Go

BERT library for Go is designed for simple data exchange between Go and Erlang/Elixir applications.

BERT stands for Binary ERlang Term. It is a Remote Procedure Call mechanism to support interop between Erlang code\
and other programming languages.

BERT library implements serialization and deserialization, as well as a subset / variation of the BERT-RPC protocol
for Go.

Here are the important points to note:
- This version supports BERT-RPC over HTTP. It is not optimal for performance, but can rely on standard HTTP tooling
  features like connection pools, authentication, load balancing, etc.
- This version implements the type I needed in Erlang External Term Format for interop with
  [ejabberd](https://github.com/processone/ejabberd/).
  
### Why use BERT?

If you want to exchange data with Erlang node, it is handy to use a format that support all the Erlang types, including
atoms. Without having the concept of atoms explicitly in the data exchange protocol, you end up adding wrapper tuples
and conversions on the Erlang side that become very painful.

If you do not need to interop with Erlang, we would recommend using Protobuf or MsgPack.

## TODO

Support various transport for BERT-RPC client:
- HTTP
- TCP/IP
- MQTT
