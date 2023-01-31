# HTP

HTP uses time information present in HTTP headers to determine the clock offset
between two machines. It can be used to synchronize the local clock using a
trusted HTTP(S) server, or to determine the time of the remote machine.

> **Note**: If NTP is available, use it instead. I developed this application
> to be able to synchronize my computer's clock in a network that blocked NTP
> packets.

## Installation

Download the binary for your platform from the [releases page](https://github.com/danroc/htp/releases/latest).

Or install it with `go`:

```bash
go install github.com/danroc/htp/cmd/htp@latest
```

## Building

To build the main application, run:

```console
go build ./cmd/htp
```

## Algorithm

Suppose that _T_ is the correct time (remote time) and our local time is offset
by _θ_ (thus local time is _T + θ_).

To approximate _θ_, we perform these steps:

1. (A, local) sends a request to (B, remote) at _t₀_ (local clock)
2. (B) receives and answers (A)'s request at _t₁_ (remote clock)
3. (A) receives (B)'s answer at _t₂_ (local clock)

These steps are represented in the following diagram:

```text
            t₁
(B) --------^--------> T
           / \
          /   \
(A) -----^-----v-----> T + θ
         t₀    t₂
```

Bringing _t₁_ to the local time (between _t₀_ and _t₂_):

t₀ < t₁ + θ < t₂ ⇒ t₀ - t₁ < θ < t₂ - t₁

So,

- θ > t₀ - t₁
- θ < t₂ - t₁

But we must use _⌊t₁⌋_ instead of _t₁_ in our calculations because it is the
only time information present in the HTTP response header.

Since _t₁ ∈ [⌊t₁⌋, ⌊t₁⌋ + 1)_, then:

- θ > t₀ - ⌊t₁⌋ - 1
- θ < t₂ - ⌊t₁⌋

Observe that the closer _t₁_ is to _⌊t₁⌋_ or _⌊t₁⌋ + 1_, smaller is the error
in the second or first equation above, respectively.

We can repeat the above procedure to improve our estimate of _θ_:

- θ⁻ = MAX(θ⁻, t₀ - ⌊t₁⌋ - 1)
- θ⁺ = MIN(θ⁺, t₂ - ⌊t₁⌋)
- θ = (θ⁺ + θ⁻)/2

The ideal delay _d_ to wait before sending the next request is calculated so
that the next value of _t₁_ is close to a "full" second:

t₂ + d + (t₂ - t₀)/2 - θ = ⌊t₁⌋ + k, k ∈ ℤ

⇒ d = ⌊t₁⌋ + k + θ - t₂ - (t₂ - t₀)/2 mod 1

⇒ d = θ - t₂ - (t₂ - t₀)/2 mod 1

Where:

- _(t₂ - t₀)_ is an estimation of the round-trip time (RTT).
- _- θ_ converts from local to remote time.

## Alternatives

- One-liner:

  ```bash
  date -s "$(curl -fsSLI https://www.google.com | grep -i "Date: " | cut -d" " -f2-)"
  ```

- [Time over HTTPS](http://phk.freebsd.dk/time/20151129/)

- [htpdate](https://www.vervest.org/htp/)
