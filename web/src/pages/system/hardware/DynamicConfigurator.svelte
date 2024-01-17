<script lang="ts">
    import type { ConfigSpec, ConfigSpecInner, ConfigSpecStruct } from "../driver";
    import { MDCTextField } from '@material/textfield';
    // import '@material/web/icon/icon.js';
    // import '@material/web/iconbutton/filled-icon-button.js';
    // import '@material/web/iconbutton/filled-tonal-icon-button.js';
    // import '@material/web/iconbutton/icon-button.js';
    // import '@material/web/iconbutton/outlined-icon-button.js';
    import {MDCRipple} from '@material/ripple';

    type HtmlInputType = 'number' | 'text';

    export let topLevelLabel: string | null = null
    export let spec: ConfigSpec = null;
    let dom: HTMLDivElement = null

    interface HtmlTree {
        html: HTMLElement
        handle: MDCTextField | null
        source: ConfigSpec
    }

    function deleteListElement(element: HTMLElement) {
        element.remove()
    }

    function createListElement(spec: ConfigSpec): HTMLElement {
        const tree = specToHtml(spec, null);
        return tree.html
    }

    function createIconButton(mdIcon: string, callback: () => void): HTMLElement {
        let buttonOuter = document.createElement('button')
        buttonOuter.classList.add('mdc-icon-button', 'material-icons')

        let buttonRippleInner = document.createElement('div')
        buttonRippleInner.classList.add('mdc-icon-button__ripple')
        buttonOuter.appendChild(buttonRippleInner)

        buttonOuter.append(mdIcon)

        const iconButtonRipple = new MDCRipple(buttonOuter);
        iconButtonRipple.unbounded = true;

        buttonOuter.onclick = callback

        return buttonOuter
    }

    function createListConfigurator(nestedSpec: ConfigSpec, label: string | null): HTMLElement {
        let listContainer = document.createElement('div')
        listContainer.classList.add("config-option__list", "mdc-elevation--z6")

        // Create heading
        if (label !== null) {
            let listFieldName = document.createElement('span')
            listFieldName.classList.add("config-option__list__heading", "text-hint")
            listFieldName.innerText = label
            listContainer.appendChild(listFieldName)
        }

        // Create main body which contains the individual fields
        let listBody = document.createElement("li")
        listBody.classList.add("config-option__list__body")
        listContainer.appendChild(listBody)

        // Create footer button to add elements
        listContainer.appendChild(createIconButton('add', () => {
            let listElementWrapper = document.createElement('li')
            listElementWrapper.classList.add('config-option__list__body__item')

            let listElementDeleteWrapper = document.createElement('div')
            listElementDeleteWrapper.classList.add('config-option__list__body__item__delete')

            let listElementDelete = createIconButton('delete', () => {
                deleteListElement(listElementWrapper)
            })

            listElementDeleteWrapper.appendChild(listElementDelete)
            listElementWrapper.appendChild(listElementDeleteWrapper)

            listElementWrapper.appendChild(createListElement(nestedSpec))
            listBody.appendChild(listElementWrapper)
        }))

        return listContainer
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
                const listSpec = spec as ConfigSpecInner
                console.error("TODO", spec.type)

                let listHtml = createListConfigurator(listSpec.inner, label)

                // const [html, handle] = newTextField('foobar', 'number', label)

                return {
                    html: listHtml,
                    handle: null,
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
                let structParent = document.createElement('div')
                structParent.classList.add('config-option__struct', 'mdc-elevation--z6')

                if (label !== null) {
                    let labelSpan = document.createElement('span')
                    labelSpan.classList.add("config-option__struct__label", "text-hint")
                    labelSpan.innerText = label
                    structParent.appendChild(labelSpan)
                }

                let fieldListHtml: HTMLUListElement = document.createElement('ul')
                fieldListHtml.classList.add("config-option__struct__fields") // TODO: deal with these classes

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
                    listElement.classList.add('config-option__struct__fields__field')

                    listElement.appendChild(subTree.html)
                    fieldListHtml.appendChild(listElement)
                }

                structParent.appendChild(fieldListHtml)

                // TODO: how to deal with mdc handles???
                return {
                    html: structParent,
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

    function generateInputs(data: ConfigSpec | null, topLevelLabel: string | null) {
        console.log('hi')

        if (data === null || dom === null) {
            return
        }

        const generatedTree = specToHtml(data, topLevelLabel)

        // TODO: use mdc component to customize it
        console.dir(generatedTree)

        // NOTE: Container must be cleared since reactive updates would cause repeated appensions which then causes
        // the configuration elements to be displayed multiple times.
        // TODO: Is it required / clean to somehow destroy the MDC components from JS?
        dom.innerHTML = ""
        dom.appendChild(generatedTree.html)
    }

    $: if (spec !== null) generateInputs(spec, topLevelLabel)
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

    :global(ul) {
        list-style-type: none;

        :global(li:not(:last-child)) {
            margin-bottom: .5rem;
        }
    }

    :global(.config-option__list) {
        background-color: var(--clr-height-3-6);
        padding: .5rem .8rem;
        border-radius: .3rem;
        border: .125rem solid var(--clr-height-6-12);

        :global(&__body) {
            background-color: red;

            :global(&__item) {
                background-color: red;
            }
        }
    }

    :global(.config-option__struct) {
        background-color: var(--clr-height-3-6);
        padding: .5rem .8rem;
        border-radius: .3rem;
        border: .125rem solid var(--clr-height-6-12);

        :global(&__fields) {
            :global(&__field) {

            }
        }
    }
</style>
