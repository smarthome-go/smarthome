<script lang="ts">
    import type { ConfigSpec, ConfigSpecInner, ConfigSpecStruct, ConfigSpecType } from "../driver";
    import { MDCTextField } from '@material/textfield';
    import {MDCRipple} from '@material/ripple';

    // NOTE: the output data is synced through the use of svelte events.
    // The data format is structured as primitive json, as if it was serialized from Homescript directly.

    type HtmlInputType = 'number' | 'text';

    export let data: any = {}
    $: console.dir(data)

    export let topLevelLabel: string | null = null
    export let spec: ConfigSpec = null;
    let dom: HTMLDivElement = null

    interface HtmlTree {
        html: HTMLElement
        handle: MDCTextField | null
        source: ConfigSpec
    }

    // Returns the index of the element that was removed
    function deleteListElement(childElement: HTMLElement, parentList: HTMLElement): number {
        // Determine the index of the element to be removed
        const index = getChildNodePositionInParent(childElement, parentList)
        childElement.remove()
        return index
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

    const LIST_BODY_CLASS_NAME = "config-option__list__body"
    const LIST_BODY_ITEM_CLASS_NAME = `${LIST_BODY_CLASS_NAME}__item`
    const STRUCT_FIELD_LIST_CLASS_NAME = "config-option__struct__fields"

    // TODO: patch event listeners and ids when an element is removed
    // OR: use a map which matches a parent-uri + a DOM node to a list-index
    // NOTE: For removal of elements: multiple things need to happen:
    // A new URI must be assigned to EVERY existing HTML element
    // This includes modifying the callback function, the input ID and the descriptor ID
    // Furthermore, the data object must be modified accordingly
    function createListConfigurator(nestedSpec: ConfigSpec, label: string | null, listURI: JsonUri): HTMLElement {
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
        let listBody = document.createElement("ul")
        listBody.classList.add(LIST_BODY_CLASS_NAME)
        listContainer.appendChild(listBody)

        // Create footer button to add elements
        listContainer.appendChild(createIconButton('add', () => {
            addListElement(listBody, listURI, nestedSpec)
        }))

        return listContainer
    }

    function addListElement(listBody: HTMLElement, listURI: JsonUri, nestedSpec: ConfigSpec) {
        // When an element is added, an URI change must be emitted as well.
        // Furthermore, the URIs for every list element must be recomputed.
        let listElementWrapper = document.createElement('li')
        listElementWrapper.classList.add(LIST_BODY_ITEM_CLASS_NAME)

        let listElementDeleteWrapper = document.createElement('div')
        listElementDeleteWrapper.classList.add(`${LIST_BODY_ITEM_CLASS_NAME}__delete`)

        let listElementDelete = createIconButton('delete', () => {
            // TODO: removal is going to be very expensive
            deleteListElement(listElementWrapper, listBody)

            patchListURIs(listBody, nestedSpec, listURI)

            // TODO: delete data element from data object
        })

        listElementDeleteWrapper.appendChild(listElementDelete)
        listElementWrapper.appendChild(listElementDeleteWrapper)

        // Compute the URI index for this element (used by creating a temporary replacement)
        let temporaryDiv = document.createElement('div')
        listBody.appendChild(temporaryDiv)
        const uriIndex = getChildNodePositionInParent(temporaryDiv, listBody)
        temporaryDiv.remove()

        let newUri = listURI.clone()
        newUri.push({
            type: 'index',
            listIndex: uriIndex,
            fieldName: null,
        })

        let listElementInputWrapper = document.createElement('div')
        listElementInputWrapper.classList.add(`${LIST_BODY_ITEM_CLASS_NAME}__input`)

        let listElement = specToHtml(nestedSpec, null, newUri)
        listElementInputWrapper.appendChild(listElement.html)

        listElementWrapper.appendChild(listElementInputWrapper)
        listBody.appendChild(listElementWrapper)
        
        let thisListValue = getUriValue(listURI) as []
        thisListValue.push(createDefaultDataFromSpec(spec))
    }

    function createDefaultDataFromSpec(spec: ConfigSpec): any {
        switch (spec.type) {
            case 'INT':
            case 'FLOAT':
            case 'BOOL':
            case 'STRING':
            case 'LIST':
            case 'STRUCT':
            case 'OPTION':
            default:
                throw(`A new config spec was introduced without updating this code: ${spec.type}`)
        }
    }

    // TODO: remove this
    // This function uses the nested spec to traverse the children of the list recursively,
    // updating the URIs to use the correct indices.
    // This function is only used when a list element is removed.
    // function patchListChildURIs(
    //     bodyNode: HTMLElement,
    //     nestedSpec: ConfigSpec,
    //     listURI: JsonUri,
    // ) {
    //     for (let child of bodyNode.children) {
    //         patchElementURI(child as HTMLElement, nestedSpec, listURI)
    //     }
    // }

    //
    // URI DOM node patching
    //

    function patchElementURI(thisNode: HTMLElement, typeSpec: ConfigSpec, parentURI: JsonUri) {
        switch (typeSpec.type) {
            case 'INT':
            case 'FLOAT':
            case 'BOOL':
            case 'STRING':
                patchTextField(thisNode, parentURI)
                break
            case 'LIST':
                const listSpec = typeSpec as ConfigSpecInner
                patchListURIs(thisNode, listSpec.inner, parentURI)
                break
            case 'STRUCT':
                const structSpec = typeSpec as ConfigSpecStruct
                patchStructURIs(thisNode, structSpec, parentURI)
                break
            case 'OPTION':
                // TODO: handle wrapper element?
                const optionSpec = typeSpec as ConfigSpecInner
                patchOptionUri(thisNode, optionSpec.inner, parentURI)
                break
            default:
                console.dir(spec)
                console.error(`BUG warning: a new spec type (${spec.type}) was introduced without updating this code`)
                break
        }
    }

    function patchListURIs(listParent: HTMLElement, listInnerTypeSpec: ConfigSpec, parentURI: JsonUri) {
        console.log('patching list uris...')
        console.dir(listParent)

        let listChildren = listParent.children
        if (listChildren.length === 0) {
            // No children => no work todo
            return
        }

        for (let idx = 0; idx < listChildren.length; idx++) {
            let listChild = listChildren[idx]
            const query = `.${LIST_BODY_ITEM_CLASS_NAME}__input`
            let inputNode = listChild.querySelector(query)
            if (inputNode === null) {
                console.dir(listChild)
                throw(`Cannot get input element of list element with query ${query}`)
            }

            let newUri = parentURI.clone()
            newUri.push({
                type: 'index',
                fieldName: null,
                listIndex: idx,
            })

            console.log(`Assigned new uri: ${newUri.string()}`)

            patchElementURI(inputNode as HTMLElement, listInnerTypeSpec, newUri)
            console.log(`patched idx: ${idx}`)
        }
    }

    function patchStructURIs(structParent: HTMLElement, structTypeSpec: ConfigSpecStruct, parentURI: JsonUri) {
        let fieldListHtml = structParent.querySelector(`.${STRUCT_FIELD_LIST_CLASS_NAME}`)

        let fieldList = fieldListHtml.children
        if (fieldList.length === 0) {
            // No fields => bug
            throw("Struct contains no fields")
        }

        for (let idx = 0; idx < fieldList.length; idx++) {
            let fieldType = structTypeSpec.fields[idx]

            let liElement = fieldList[idx] as HTMLElement

            let inputNode = liElement.firstChild
            if (inputNode === null) {
                throw("Cannot get input node in list children")
            }

            patchElementURI(inputNode as HTMLElement, fieldType.type, parentURI)
        }

    }

    function patchOptionUri(optionParent: HTMLElement, innerTypeSpec: ConfigSpec, parentURI: JsonUri) {
        throw("Option patching is not implemented yet")
    }

    //
    // END URI patching
    //

    function getChildNodePositionInParent(child: HTMLElement, parent: HTMLElement) : number {
        for (let idx = 0; idx < parent.children.length; idx ++) {
            // Test for equality
            if (child.isEqualNode(parent.children.item(idx))) {
                return idx
            }
        }

        throw "Passed child node parameter is not an actal child of the parent parameter"
    }

    interface JsonUriComponent {
        type: 'field' | 'index'
        fieldName: string | null
        listIndex: number | null
    }

    class JsonUri {
        fields: JsonUriComponent[]

        constructor(fromStr?: string) {
            this.fields = []

            if (!fromStr) {
                return
            }

            // Split at `::`
            let chunks = fromStr.split("::")

            for (let chunk of chunks) {
                let chunkSplit = chunk.split("@")
                switch (chunkSplit[0]) {
                    case "field":
                        this.fields.push({
                            type: 'field',
                            fieldName: chunkSplit[1],
                            listIndex: null,
                        })
                        break
                    case "idx":
                        this.fields.push({
                            type: 'index',
                            fieldName: null,
                            listIndex: parseInt(chunkSplit[1]),
                        })
                        break
                    default:
                        throw(`Unsupported ID encoding chunk: ${chunk}`)
                }
            }
        }

        push(newComponent: JsonUriComponent) {
            this.fields.push(newComponent)
        }

        pop(): JsonUriComponent {
            return this.fields.pop()
        }

        popFront(): JsonUriComponent {
            return this.fields.shift()
        }

        isEmpty(): boolean {
            return this.fields.length === 0
        }

        parent(): JsonUri {
            let cloned = this.clone()
            cloned.pop()
            return cloned
        }

        // If parent is n segments long, the first n segments of this URI will be changed to the one of the parent.
        overrideStart(parent: JsonUri) {
            let parentCloned = parent.clone()

            // Overwrite the first n segments
            for (let idx = 0; idx < parentCloned.fields.length; idx++) {
                this.fields[idx] = parentCloned.fields[idx]
            }
        }

        clone(): JsonUri {
            let newUri = new JsonUri()

            for (let field of this.fields) {
                newUri.push({
                    type: field.type,
                    fieldName: field.fieldName,
                    listIndex: field.listIndex,
                })
            }

            return newUri
        }

        string(): string {
            return this.fields.map((field) => {
                switch (field.type) {
                    case "field":
                        // TODO: sanitize field name: base64 encoding?
                        return `field@${field.fieldName}`
                    case "index":
                        return `idx@${field.listIndex}`
                    default:
                        throw(`BUG warning: a new URI component type was added without updating this code (${field.type})`)
                }
            }).join("::")
        }
    }

    function getUriValue(uri: JsonUri): any {
        console.log(`Getting data from ${uri}...`)

        // Traverse and manipulate
        let addressableData = data

        let lastSegment = null

        while (1) {
            lastSegment = uri.pop()

            if (uri.fields.length === 0) {
                // Perform assignment, address has been reached

                switch (lastSegment.type) {
                    case 'field':
                        return addressableData[lastSegment.fieldName]
                    case 'index':
                        return addressableData[lastSegment.listIndex]
                    default:
                        throw(`A new segment type was introduced without updating this code`)
                }
            }

            switch (lastSegment.type) {
                case 'field':
                    addressableData = addressableData[lastSegment.fieldName]

                    if (addressableData !== null) {
                        throw(`field ${lastSegment.fieldName} is undefined`)
                    }

                    break
                case 'index':
                    addressableData = addressableData[lastSegment.listIndex]

                    if (addressableData !== null) {
                        throw(`index ${lastSegment.listIndex} is undefined`)
                    }

                    break
                default:
                    throw(`A new segment type was introduced without updating this code`)
            }
        }

        return addressableData
    }

    function maniplateUriValue(uri: JsonUri, newData: any) {
        console.log(`Manipulating ${uri} to ${newData}`)

        let targetData = getUriValue(uri)
        targetData = newData

        return

        // Traverse and manipulate
        let addressableData = data

        let lastSegment = null

        while (1) {
            lastSegment = uri.pop()

            if (uri.fields.length === 0) {
                // Perform assignment, address has been reached

                switch (lastSegment.type) {
                    case 'field':
                        addressableData[lastSegment.fieldName] = newData
                        break
                    case 'index':
                        addressableData[lastSegment.listIndex] = newData
                        break
                    default:
                        throw(`A new segment type was introduced without updating this code`)
                }

                break
            }

            switch (lastSegment.type) {
                case 'field':
                    addressableData = addressableData[lastSegment.fieldName]

                    if (addressableData !== null) {
                        throw(`field ${lastSegment.fieldName} is undefined`)
                    }

                    break
                case 'index':
                    addressableData = addressableData[lastSegment.listIndex]

                    if (addressableData !== null) {
                        throw(`index ${lastSegment.listIndex} is undefined`)
                    }

                    break
                default:
                    throw(`A new segment type was introduced without updating this code`)
            }
        }

        // if (lastSegment === null) {
        //     throw(`Cannot manipulate URI value: last segment is null`)
        // }
    }

    function specToHtml(spec: ConfigSpec, label: string | null, uri: JsonUri): HtmlTree {
        console.log('specToHtml: ', spec, label)

        switch (spec.type) {
            case 'INT': {
                const [html, handle] = newTextField(uri, 'number', label)

                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'BOOL': {
                console.error("TODO", spec.type)

                const [html, handle] = newTextField(uri, 'number', label)

                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'LIST': {
                const listSpec = spec as ConfigSpecInner
                console.error("TODO", spec.type)

                let listHtml = createListConfigurator(listSpec.inner, label, uri)

                return {
                    html: listHtml,
                    handle: null,
                    source: spec,
                }
            }
            case 'FLOAT': {
                const [html, handle] = newTextField(uri, 'number', label)

                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'STRING': {
                const [html, handle] = newTextField(uri, 'text', label)

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
                    // Add the name to the text element URI
                    let newURI = uri.clone()
                    newURI.push({
                        type: 'field',
                        fieldName: field.name,
                        listIndex: null,
                    })

                    const subTree = specToHtml(field.type, field.name, newURI)
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

    function onInputHook(uri: JsonUri, inputElement: HTMLInputElement){
            console.log(`input "${uri.string()}" changed to ${inputElement.value}`)

            // TODO: implement string / bool -> string parsing
            maniplateUriValue(uri, inputElement.value)
    }

    function newTextField(uri: JsonUri, inputType: HtmlInputType, labelText: string | null): [HTMLElement, MDCTextField] {
        const labelId = uri.string()

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

        // Listen to update events
        // TODO: persist these changes via URI addressing
        inputElement.oninput = (_) => { onInputHook(uri, inputElement) }

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

    function patchTextField(textFieldParent: HTMLElement, parentURI: JsonUri) {
        let inputInnerLabel = textFieldParent.querySelector('.mdc-floating-label')
        if (inputInnerLabel === null) {
            throw("Input inner label of textfield not found")
        }

        // Parse old URI
        const oldUri = new JsonUri(inputInnerLabel.id)
        // Create new URI
        let newUri = oldUri.clone()
        newUri.overrideStart(parentURI)

        console.log(`old uri: ${oldUri.string()} | new uri: ${newUri.string()} | patch fragment: ${parentURI.string()}`)

        const newUriId = newUri.string()
        // Update id of label
        inputInnerLabel.id = newUriId

        let inputElement: HTMLInputElement = textFieldParent.querySelector(".mdc-text-field__input")
        if (inputElement === null) {
            throw("Input element of textfield not found")
        }
        // Update reference to label id
        inputElement.setAttribute('aria-labelledby', newUriId)

        // Update on-input hook
        inputElement.oninput = (_) => { onInputHook(newUri, inputElement) }

        console.log("patched text field")
    }

    function generateInputs(data: ConfigSpec | null, topLevelLabel: string | null) {
        console.log('hi')

        if (data === null || dom === null) {
            return
        }

        const generatedTree = specToHtml(data, topLevelLabel, new JsonUri())

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
