function addLoadEvent(func: () => void) {
    const oldOnLoad: any = window.onload
    if (typeof window.onload != 'function') {
        window.onload = func
    } else {
        window.onload = function () {
            if (oldOnLoad) {
                oldOnLoad(undefined)
            }
            func()
        }
    }
}