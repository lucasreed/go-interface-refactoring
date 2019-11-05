Refactoring to use go interfaces
===

NOTE: the `after-ednsctl` module doesn't do anything when you run main.go right now, even when provided the proper "required" arguments. This module is to show a skeleton of the implementation I have been thinking up. Primarily this work has just been in the `pkg/dns` package and the specific implementations that are present under that directory

Before
---

In the `before-ednsctl` module, the primary package you will want to look at is `pkg/internal/dns`. It contains a lot of copy/pasted code to be used for specific DNS providers that could be abstracted a bit better.

After
---

In the `after-ednsctl` module, the primary package you will want to compare with the before module is `pkg/dns` and then all the other provider-specific packages at implement the `dns.API` Interface.
