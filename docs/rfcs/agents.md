# Agents

This document provides an overview and comparison of several libraries that
could be used to implement agent pipelines.

## Reviewed libraries

The following options are slated for review:

- [charm.land/fantasy]( https://pkg.go.dev/charm.land/fantasy) ([github
  repo](https://github.com/charmbracelet/fantasy))
- [Protocol-Lattice/go-agent](https://pkg.go.dev/github.com/Protocol-Lattice/go-agent)
  from [Protocol Lattice](https://github.com/Protocol-Lattice) organization
  ([github repo](https://github.com/Protocol-Lattice/go-agent))

According to popularity statistics, *fantasy* currently has more GitHub stars
than *go‑agent*, though the numbers may be inflated by hype. While
**[charm.land](https://charm.land)** boasts a large community,
**[Protocol‑Lattice](https://github.com/Protocol-Lattice)** appears to
prioritize performance. We’ll delve into this contrast in our review.
Nonetheless, it’s interesting to track how GitHub star counts change over time.

[![Star History Chart](https://api.star-history.com/svg?repos=charmbracelet/fantasy,Protocol-Lattice/go-agent&type=date&legend=top-left)](https://www.star-history.com/#charmbracelet/fantasy&Protocol-Lattice/go-agent&type=date&legend=top-left)
