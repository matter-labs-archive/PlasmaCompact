# Some info

Our Move Viable Plasma implementation is live at [plasma-mainnet.thematter.io](https://plasma-mainnet.thematter.io) with block explorer [here](https://plasma-explorer-mainnet.thematter.io) with verifier and mobile client is coming in the next weeks. Mobile client work is happening [here](https://github.com/matterinc/DiveLane), please help if you want.

## Description

Overnight implementation of ERDAX proposed block structure for UTXO model where block only requires the block itself and the previous header. A detailed explaier is below.

## Disclaimer

This is a code written overnight during the Devcon 4, when I had a second thought about viability of such block and transaction structure for Plasma implementations, especially Move Viable Plasma. This code is not intended for production.

## What is Plasma and their trade-offs

One should understand that Plasma (as a philisophy) itself is a trade-off where speed comes at the cost of requirement for continuous (or periodic) verification and potentially of sophisticated exit game.

Here is a short comparison of different flavours of Plasma.

|               | MoreVP (or MVP)                                                           | Plasma Cash                             | Plasma Compact                                                                                                                                           |
|---------------|---------------------------------------------------------------------------|-----------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| What you get  | Speed (due to UTXO concurrency)                                           | Ease of exits                           | Ease of verification (only block itself and previous header is required)                                                                                 |
| At what price | - Verification - Storage (UTXO index size) - Joint exit game (exit queue) | - Storage (coin history) - Non-fungible | - Bandwidth (block sizes and transaction sizes are larger) - Speed penalty (SMT update round is added) - Proof updates on the client (can be outsourced) |

While all them are viable for their purposes, at the current state of technology (and UX is an important part of it) we should consider bandwidth to be cheaper than storage taken a large amount of mobile devices. Of course, an ideal implementation will be account model Plasma, but it has it's own challenges (stay tuned for a soon announcement!).

## Authors

- Alex Vlasov, [@shamatar](https://github.com/shamatar)

