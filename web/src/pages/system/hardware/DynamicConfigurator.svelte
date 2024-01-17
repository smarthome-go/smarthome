<script lang="ts">
    import type { ConfigSpec, ConfigSpecStruct } from "../driver";
    import { MDCTextField } from '@material/textfield';

    type HtmlInputType = 'number' | 'text';

    export let spec: ConfigSpec = null;
    let dom: HTMLDivElement = null

    interface HtmlTree {
        html: HTMLElement
        handle: MDCTextField | null
        source: ConfigSpec
    }

    function specToHtml(spec: ConfigSpec, label: string | null): HtmlTree {
        console.log('specToHtml: ', spec, label)

        switch (spec.type) {
            case 'INT': {
                const [html, handle] = newTextField('foobar', 'number', label)

                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'BOOL': {
                console.error("TODO", spec.type)

                const [html, handle] = newTextField('foobar', 'number', label)

                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'LIST': {
                console.error("TODO", spec.type)

                const [html, handle] = newTextField('foobar', 'number', label)

                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'FLOAT': {
                const [html, handle] = newTextField('foobar', 'number', label)

                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'STRING': {
                const [html, handle] = newTextField('foobar', 'text', label)

                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'STRUCT':
                let fieldListHtml: HTMLUListElement = document.createElement('ul')
                fieldListHtml.classList.add("config-option__field-list") // TODO: deal with these classes

                let fields = (spec as ConfigSpecStruct).fields

                // Display a special message if there are no fields
                if (fields.length === 0) {
                    // The only case where this message is ever shown is if the driver has no configration parameters.
                    let isEmptyMessage = document.createElement('span')
                    isEmptyMessage.classList.add("driver-no-configuration")
                    isEmptyMessage.innerText = "Driver does not have any configuration parameters."

                    return {
                        html: isEmptyMessage,
                        handle: null,
                        source: spec,
                    }
                }

                for (let field of fields) {
                    const subTree = specToHtml(field.type, field.name)
                    let listElement = document.createElement('li')
                    listElement.appendChild(subTree.html)
                    fieldListHtml.appendChild(listElement)
                }

                // TODO: how to deal with mdc handles???
                return {
                    html: fieldListHtml,
                    handle: null,
                    source: spec,
                }
            case 'OPTION':
                break
            default:
                console.dir(spec)
                console.error(`BUG warning: a new spec type (${spec.type}) was introduced without updating this code`)
                return {
                    html: document.createElement('span'),
                    handle: null,
                    source: spec,
                }
        }
    }

    function newTextField(labelId: string, inputType: HtmlInputType, labelText: string | null): [HTMLElement, MDCTextField] {
        let inputInnerLabel = document.createElement('span')
        inputInnerLabel.classList.add("mdc-floating-label")
        inputInnerLabel.id = labelId
        if (labelText !== null) {
            inputInnerLabel.innerText = labelText
        }

        let inputElement = document.createElement('input')
        inputElement.type = inputType
        inputElement.classList.add("mdc-text-field__input")
        inputElement.setAttribute('aria-labelledby', labelId)

        let rippleSpan = document.createElement('span')
        rippleSpan.classList.add('mdc-text-field__ripple')

        let lineRipple = document.createElement('span')
        lineRipple.classList.add('mdc-line-ripple')

        let outerLabel = document.createElement('label')
        outerLabel.classList.add('mdc-text-field', 'mdc-text-field--filled')

        outerLabel.append(rippleSpan, inputElement, inputInnerLabel, lineRipple)
        const textField = new MDCTextField(outerLabel); // TODO: check if this works

        return [outerLabel, textField]
    }

    function generateInputs(data: ConfigSpec | null) {
        console.log('hi')

        if (data === null || dom === null) {
            return
        }

        const generatedTree = specToHtml(data, "TODO: label")

        // TODO: use mdc component to customize it
        console.dir(generatedTree)

        // NOTE: Container must be cleared since reactive updates would cause repeated appensions which then causes
        // the configuration elements to be displayed multiple times.
        // TODO: Is it required / clean to somehow destroy the MDC components from JS?
        dom.innerHTML = ""
        dom.appendChild(generatedTree.html)
    }

    $: if (spec !== null) generateInputs(spec)
</script>

<div class="configurator">
    {#if spec !== null}
        <div class="configurator__dom" bind:this={dom}></div>
    {/if}
</div>

<style lang="scss">
    @use "@material/floating-label/mdc-floating-label";
    @use "@material/line-ripple/mdc-line-ripple";
    @use "@material/notched-outline/mdc-notched-outline";
    @use "@material/textfield";
    @include textfield.core-styles;

</style>
