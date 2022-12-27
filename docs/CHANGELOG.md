## Changelog for v0.6.0

### Additions

- Added the `remind` builtin function to Homescript
- The `remind` function adds a new reminder to the user's reminders
- It can be used like this:

```py
remind(
   'title',
   'description',
   1, // 1 - 5 (urgency)
   {
     day: time.now().calendar_day,
     month: time.now().month,
     year: time.now().year,
   }
 );
```

### Bugfixes

- Fixed minor bugs in Homescript
