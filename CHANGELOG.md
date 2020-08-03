## Changes

### 0.2.0

- Swap out the workerpool with gammazero workerpool implementation
- Changed interface of `DoBunch` to now return an error for when the client has been closed and no further requests can be made
- Added `.Destroy()` function to close all channels and workers that the client has created, it will return after all work has been done