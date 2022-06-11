## Changelog for v0.0.42

### Homescript GUI
- Added a better argument display
- Made the `AddArgument` dialog non-fullscreen due to unevent proportions
- Introduced data-type and display selectors in the `EditArgument` dialog

### Homescript
- Implemented a call stack
- Using the call stack, Homescript can now detect script recursion
- When script recursion is detected, the execution is prevented and the user is informed about this event

#### Example code using `v0.0.41`
As seen below, the script calls itself indefinitely, causing endless recursion
*Input*: `test.hms`
```python
# Current script: test
exec('test')
```
*Output*: Server would run out of resources, database would return errors, however no complete crash

#### Example code using `v0.0.42`
As seen below, the scripts calls themselves sequentially without ending, causing endless recursion
*Input*: `start.hms`, `first.hms`, `second.hms`
```python
# Current script: start
exec('first')
```
```python
# Current script: first
exec('second')
```
```python
# Current script: second
exec('start')
```

*Output*: Homescripts detects recursion
```
Error: Program terminated abnormally with exit-code 1
RuntimeError at start.hms:2:1

  1  | # Write your code for `start` below
  2  | exec('first')
       ^
  3  |

Homescript terminated with exit code 1: Homescript terminated with exit code 1: Exec violation: executing 'start' could cause infinite recursion.
=== Call Stack ===
   0: start      (INITIAL)
   1: first
   2: second
   3: start      (PREVENTED)

Homescript terminated with exit code: 1 [0.89s]
```
##### Limitations
- The recursion detector only checks if the script to-be-executed is already present in the call stack
- Due to this, recursion is completely forbidden
- However, recursion is not meant to be used in Homescript, therefore not being a limitation
