# HTP

HTP uses time information present in HTTP headers to determine the clock offset
between two machines. It can be used to synchronize the local clock using a
trusted HTTP(S) server, or to determine the time of the remote machine.

## Build

```console
go build ./cmd/htp
```

## Installation

Download the binary from the [releases page](https://github.com/danroc/htp/releases/latest).

## Algorithm

Suppose that $T$ is the correct time (remote time) and our local time is offset
by $\theta$.

To approximate $\theta$, we perform these steps:

1. (A) sends a request to (B) at $t_0$ (local clock)
2. (B) receives and answers (A)'s request at $t_1$ (remote clock)
3. (A) receives (B)'s answer at $t_2$ (local clock)

These steps are represented in the following diagram:

```text
            t₁
(B) --------^--------> T
           / \
          /   \
(A) -----^-----v-----> T + θ
         t₀    t₂
```

Bringing $t_1$ to the local time (between $t_0$ and $t_2$):

$$
t_0 < t_1 + \theta < t_2 \Rightarrow t_0 - t_1 < \theta < t_2 - t_1
$$

So,

- $\theta > t_0 - t_1$
- $\theta < t_2 - t_1$

But we must use $\lfloor t_1 \rfloor$ instead of $t_1$ in our calculations
because it is the only time information present in the HTTP response header.

Since $t_1 \in [\lfloor t_1 \rfloor, \lfloor t_1 \rfloor + 1)$, then:

- $\theta > t_0 - \lfloor t_1 \rfloor - 1$
- $\theta < t_2 - \lfloor t_1 \rfloor$

Observe that the closer $t_1$ is to $\lfloor t_1 \rfloor$ or $\lfloor t_1
\rfloor + 1$, smaller is the error in the second or first equation above,
respectively.

We can repeat the above procedure to improve our estimate of $\theta$:

- $\theta^- = MAX(\theta^-, t_0 - \lfloor t_1 \rfloor - 1)$
- $\theta^+ = MIN(\theta^+, t_2 - \lfloor t_1 \rfloor)$
- $\theta = (\theta^+ + \theta^-)/2$

The ideal delay $d$ to wait before sending the next request is calculated so
that the next value of $t_1$ is close to a "full" second:

$$
t_2 + d + (t_2 - t_0)/2 - \theta = \lfloor t_1 \rfloor + k, k \in \mathbb{Z}\\
\Rightarrow d = \lfloor t_1 \rfloor + k + \theta - t_2 - (t_2 - t_0)/2 \mod 1\\
\Rightarrow d = \theta - t_2 - (t_2 - t_0)/2 \mod 1
$$

Where:

- $(t_2 - t_0)$ is an estimation of the round-trip time (RTT).
- $-\theta$ converts from local to remote time.
