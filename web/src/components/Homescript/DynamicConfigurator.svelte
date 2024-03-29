<script lang="ts">
    import type { ConfigSpec, ConfigSpecInner, ConfigSpecStruct, ConfigSpecType } from "../../driver";
    import { MDCTextField } from '@material/textfield';
    import { MDCRipple } from '@material/ripple';
    // import { MDCMenu } from '@material/menu';
    import { MDCSelect } from '@material/select';

    import { createEventDispatcher, onMount } from 'svelte'

    const dispatch = createEventDispatcher()

    type HtmlInputType = TextInputType | 'boolean';
    type TextInputType = 'number' | 'text';

    // TODO: write comment
    export let inputData: any = null

    // TODO: if data changes externally, the contents of the inputs should also be redrawn

    // Any UI inputs will first be committed to this data structure.
    // Afterwards, the changes are synced to the external data object in order to trigger Svelte's reactive update.
    let dataInternal: any = {}

    export let topLevelLabel: string | null = null
    export let spec: ConfigSpec = null;
    let dom: HTMLDivElement = null

    interface HtmlTree {
        html: HTMLElement
        handle: MDCTextField | MDCSelect | null
        source: ConfigSpec
    }

    interface MDCSelectWrapper {
        select: MDCSelect,
        listener: () => void,
        uri: JsonUri,
    }

    let boolDropDownStore: MDCSelectWrapper[] = []

    // Returns the index of the element that was removed
    function deleteListElement(childElement: HTMLElement, parentList: HTMLElement): number {
        // Determine the index of the element to be removed
        const index = getChildNodePositionInParent(childElement, parentList)
        console.log(`Removing index ${index} from list`)
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
    function createListConfigurator(
        nestedSpec: ConfigSpec,
        label: string | null,
        listURI: JsonUri,
        currentValues: any[],
    ): HTMLElement {
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

        for (let element of currentValues) {
            addListElement(listBody, listURI, nestedSpec, element, false)
        }

        // Create footer button to add elements
        listContainer.appendChild(createIconButton('add', () => {
            addListElement(listBody, listURI, nestedSpec, createDefaultDataFromSpec(nestedSpec), true)
        }))

        return listContainer
    }

    function addListElement(
        listBody: HTMLElement,
        listURI: JsonUri,
        nestedSpec: ConfigSpec,
        currentValue: any,
        // NOTE: this is required so that a default value is not added by accident
        userClickedButton: boolean,
    ) {
        // When an element is added, an URI change must be emitted as well.
        // Furthermore, the URIs for every list element must be recomputed.
        let listElementWrapper = document.createElement('li')
        listElementWrapper.classList.add(LIST_BODY_ITEM_CLASS_NAME)

        let listElementDeleteWrapper = document.createElement('div')
        listElementDeleteWrapper.classList.add(`${LIST_BODY_ITEM_CLASS_NAME}__delete`)

        let listElementDelete = createIconButton('delete', () => {
            // NOTE: removal is quite expensive
            const removalIndex = deleteListElement(listElementWrapper, listBody)

            patchListURIs(listBody, nestedSpec, listURI)

            // Deleting this list element from the underlying data object
            getUriValue(listURI.clone()).splice(removalIndex, 1)
            commitState()
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

        let listElement = specToHtml(nestedSpec, null, newUri, currentValue)
        listElementInputWrapper.appendChild(listElement.html)

        listElementWrapper.appendChild(listElementInputWrapper)
        listBody.appendChild(listElementWrapper)

        // NOTE: this is required to prevent the addition of an erroneous default value to the internal data representation.
        if (!userClickedButton) {
            return
        }

        // NOTE: this code updates the underlying data representation to reflect the addition of an element
        // Furthermore, a change event is emitted so that parent component can react reactively
        getUriValue(listURI.clone()).push(createDefaultDataFromSpec(nestedSpec)) // THIS IS A BUG!
        commitState()
    }

    function createDefaultDataFromSpec(spec: ConfigSpec): any {
        switch (spec.type) {
            case 'INT':
                return 0
            case 'FLOAT':
                return 0.1
            case 'BOOL':
                return false
            case 'STRING':
                return ""
            case 'LIST':
                return []
            case 'STRUCT':
                let structSpec = spec as ConfigSpecStruct
                let dataObj = {}

                for (let field of structSpec.fields) {
                    dataObj[field.name] = createDefaultDataFromSpec(field.type)
                }

                return dataObj
            case 'OPTION':
                throw(`TODO: option`)
                return null
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

    function patchElementURI(thisNode: HTMLElement, typeSpec: ConfigSpec, newParentURI: JsonUri) {
        switch (typeSpec.type) {
            case 'INT':
                // TODO: add additional constraint for `number`
                patchTextField(thisNode, newParentURI, 'number')
                break
            case 'FLOAT':
                patchTextField(thisNode, newParentURI, 'number')
                break
            case 'STRING':
                patchTextField(thisNode, newParentURI, 'text')
                break
            case 'BOOL':
                patchDropDown(thisNode, newParentURI)
                break
            case 'LIST':
                const listSpec = typeSpec as ConfigSpecInner
                patchListURIs(thisNode, listSpec.inner, newParentURI)
                break
            case 'STRUCT':
                const structSpec = typeSpec as ConfigSpecStruct
                patchStructURIs(thisNode, structSpec, newParentURI)
                break
            case 'OPTION':
                // TODO: handle wrapper element?
                const optionSpec = typeSpec as ConfigSpecInner
                patchOptionUri(thisNode, optionSpec.inner, newParentURI)
                break
            default:
                console.dir(typeSpec)
                console.error(`BUG warning: a new spec type (${typeSpec.type}) was introduced without updating this code`)
                break
        }
    }

    function patchListURIs(listParent: HTMLElement, listInnerTypeSpec: ConfigSpec, parentURI: JsonUri) {
        let listChildren = listParent.children
        if (listChildren.length === 0) {
            // No children => no work todo.
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

            // console.dir(inputNode)

            patchElementURI(inputNode as HTMLElement, listInnerTypeSpec, newUri)
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


            let inputNode = null

            let children = liElement.children
            for (let child of children) {
                // NOTE: this is a boolean dropdown, extract it specially.
                if (child.classList.contains('mdc-select')) {
                    // Get correct element
                    inputNode = liElement
                    console.log("found it")
                    console.dir(inputNode)
                    break
                }
            }

            // Otherwise, use the first child
            if (inputNode === null) {
                inputNode = liElement.firstChild
            }

            if (inputNode === null) {
                throw("Cannot get input node in list children")
            }

            console.dir(inputNode)

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

        throw "Passed child node parameter is not an actual child of the parent parameter"
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

        equals(other: JsonUri): boolean {
            if (this.fields.length != other.fields.length) {
                return false
            }

            for (let i = 0; i < this.fields.length; i++) {
                const thisE = this.fields[i]
                const otherE = other.fields[i]

                if (thisE.type != otherE.type) {
                    return false
                }

                if (thisE.fieldName != otherE.fieldName) {
                    return false
                }

                if (thisE.listIndex != otherE.listIndex) {
                    return false
                }
            }

            return true
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

    function getUriValue(uriIn: JsonUri): any {
        let uri = uriIn.clone()

        // Iteratively traverse recursive data structure.
        let addressableData = dataInternal

        let firstSegment = null

        while (1) {
            if (uri.fields.length === 0) {
                return addressableData
            }

            firstSegment = uri.popFront()

            switch (firstSegment.type) {
                case 'field':
                    addressableData = addressableData[firstSegment.fieldName]

                    if (addressableData === undefined) {
                        throw(`field ${firstSegment.fieldName} is undefined`)
                    }

                    break
                case 'index':
                    const old = addressableData
                    addressableData = old[firstSegment.listIndex]

                    if (addressableData === undefined) {
                        console.dir(old)
                        throw(`Index ${firstSegment.listIndex} is undefined`)
                    }

                    break
                default:
                    throw(`A new segment type was introduced without updating this code`)
            }
        }

        return addressableData
    }

    function maniplateUriValue(uriIn: JsonUri, newData: any) {
        let uri = uriIn.clone()
        // Pop the back from the uri as we need the parent of the field, not the field itself.
        // If we would assign to the field itself, it would be a literal, making this assignment redundant.
        // => We need a object / list reference
        const literalUri = uri.pop()
        let targetDataParent = getUriValue(uri)

        // Perform field access via the parent
        switch (literalUri.type) {
            case 'field':
                targetDataParent[literalUri.fieldName] = newData
                break
            case 'index':
                targetDataParent[literalUri.listIndex] = newData
                break
            default:
                throw("A new uri kind was introduced without updating this code")
        }
    }

    function specToHtml(
        spec: ConfigSpec,
        label: string | null,
        uri: JsonUri,
        currentValue: any,
    ): HtmlTree {
        switch (spec.type) {
            case 'INT': {
                const [html, handle] = newTextField(uri, 'number', label, currentValue)
                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'BOOL': {
                const [html, handle] = newDropDown(uri, label, currentValue)

                // const [html, handle] = newTextField(uri, 'number', label, currentValue)
                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'LIST': {
                const listSpec = spec as ConfigSpecInner
                let listHtml = createListConfigurator(listSpec.inner, label, uri, currentValue)
                return {
                    html: listHtml,
                    handle: null,
                    source: spec,
                }
            }
            case 'FLOAT': {
                const [html, handle] = newTextField(uri, 'number', label, currentValue)
                return {
                    html,
                    handle,
                    source: spec,
                }
            }
            case 'STRING': {
                const [html, handle] = newTextField(uri, 'text', label, currentValue)
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

                // Display a special message if there are no fields.
                if (fields.length === 0) {
                    // The only case where this message is ever shown is if the driver has no configuration parameters.
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
                    // Add the name to the text element URI.
                    let newURI = uri.clone()
                    newURI.push({
                        type: 'field',
                        fieldName: field.name,
                        listIndex: null,
                    })

                    const subTree = specToHtml(field.type, field.name, newURI, currentValue[field.name])
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

    // TODO: deprecate this in favor of a second, proxy data object
    function commitState() {
        dispatch('change', dataInternal)
    }

    function onInputHook(uri: JsonUri, inputElement: HTMLInputElement, type: HtmlInputType) {
            const strValue = inputElement.value
            let outputValue: any = null

            switch (type) {
                case "text":
                    outputValue = strValue
                    break
                case "number":
                    outputValue = parseInt(strValue)
                    break
                default:
                    console.error(type)
            }

            maniplateUriValue(uri, outputValue)
            commitState()
    }

    // function newAtomField(
    //     uri: JsonUri,
    //     inputType: HtmlInputType,
    //     labelText: string | null,
    //     currentValue: any,
    // ) {
    //     switch (inputType) {
    //         case 'number':
    //         case 'text':
    //             newTextField(uri, inputType, labelText, currentValue)
    //             break
    //         case 'boolean':
    //             console.error("TODO: add boolean")
    //             break
    //     }
    // }

        // let innerUl = document.createElement("ul")
        // innerUl.classList.add("mdc-menu__selection-group")

        // let innerLi = document.createElement("li")
        // innerLi.classList.add("mdc-list-item")
        // innerLi.role = "menuitem"
        //
        // let innerLiRippleSpan = document.createElement("span")
        // innerLiRippleSpan.classList.add("mdc-list-item__ripple")
        //
        // let innerLiIcon = document.createElement("span")
        // innerLiIcon.classList.add("mdc-list-item__graphic", "mdc-menu__selection-group-icon")
        // innerLiIcon.innerText = "code"
        //
        // let innerLiContent = document.createElement("span")
        // innerLiContent.classList.add("mdc-list-item__text")
        // innerLiContent.innerText = "OPTION X"
        //
        // let mdcListDivider = document.createElement("li")
        // mdcListDivider.classList.add("mdc-list-divider")
        // mdcListDivider.role = "separator"
        //
        // innerLi.appendChild(innerLiRippleSpan)
        // innerLi.appendChild(innerLiIcon)
        // innerLi.appendChild(innerLiContent)
        // innerUl.appendChild(innerLi)
        // liElement.appendChild(innerUl)
        // mdcUlList.appendChild(liElement)
        // mdcUlList.appendChild(mdcListDivider)


// <div class="mdc-menu mdc-menu-surface" id="demo-menu">
//   <ul class="mdc-list" role="menu" aria-hidden="true" aria-orientation="vertical" tabindex="-1">
//     <li>
//       <ul class="mdc-menu__selection-group">
//         <li class="mdc-list-item" role="menuitem">
//           <span class="mdc-list-item__ripple"></span>
//           <span class="mdc-list-item__graphic mdc-menu__selection-group-icon">
//             ...
//           </span>
//           <span class="mdc-list-item__text">Single</span>
//         </li>
//         <li class="mdc-list-item" role="menuitem">
//           <span class="mdc-list-item__ripple"></span>
//           <span class="mdc-list-item__graphic mdc-menu__selection-group-icon">
//            ...
//           </span>
//           <span class="mdc-list-item__text">1.15</span>
//         </li>
//       </ul>
//     </li>
//     <li class="mdc-list-divider" role="separator"></li>
//     <li class="mdc-list-item" role="menuitem">
//       <span class="mdc-list-item__ripple"></span>
//       <span class="mdc-list-item__text">Add space before paragraph</span>
//     </li>
//     ...
//   </ul>
// </div>

    // https://github.com/hperrin/svelte-material-ui/tree/v7/packages/select
    function newDropDown(
        uri: JsonUri,
        labelText: string | null,
        currentValue: boolean,
    ): [HTMLElement, MDCSelect] {
        let body = document.createElement('div')
        body.classList.add("mdc-select", "mdc-select--filled", "mdc-select")
        body.id = uri.string()

        body.innerHTML = `
        <div class="mdc-select__anchor">
            <span class="mdc-select__ripple"></span>
            <span class="mdc-floating-label mdc-floating-label--float-above">${labelText !== null ? labelText : ""}</span>
            <span class="mdc-select__selected-text-container">
            <span class="mdc-select__selected-text">${labelText !== null ? labelText : ""}</span>
            </span>
            <span class="mdc-select__dropdown-icon">
                <svg
                    class="mdc-select__dropdown-icon-graphic"
                    viewBox="7 10 10 5" focusable="false">
                    <polygon
                        class="mdc-select__dropdown-icon-inactive"
                        stroke="none"
                        fill-rule="evenodd"
                        points="7 10 12 15 17 10">
                    </polygon>
                    <polygon
                        class="mdc-select__dropdown-icon-active"
                        stroke="none"
                        fill-rule="evenodd"
                        points="7 15 12 10 17 15">
                    </polygon>
                </svg>
            </span>
            <span class="mdc-line-ripple"></span>
        </div>

        <div class="mdc-select__menu mdc-menu mdc-menu-surface">
            <ul class="mdc-deprecated-list">
                <li class="mdc-deprecated-list-item" data-value="true">
                    <span class="mdc-deprecated-list-item__ripple"></span>
                    <span class="mdc-deprecated-list-item__text">True</span>
                </li>
                <li class="mdc-deprecated-list-item mdc-deprecated-list-item--selected" data-value="false" aria-selected="true">
                    <span class="mdc-deprecated-list-item__ripple"></span>
                    <span class="mdc-deprecated-list-item__text">False</span>
                </li>
            </ul>
        </div>`

        const select = new MDCSelect(body);
        const listener = registerMDCSelectOnChange(select, uri.clone(), null)
        boolDropDownStore = [...boolDropDownStore, {
            select,
            uri: uri.clone(),
            listener,
        }]

        select.setValue(`${currentValue}`, true)

        return [body, select]
    }

    function registerMDCSelectOnChange(select: MDCSelect, uri: JsonUri, oldClosure: (() => void) | null): () => void {
        const newClosure = () => {
            console.log(`Selected option at index ${select.selectedIndex} with value "${select.value} | URI: ${uri.string()}"`);

            let boolOut = null

            switch (select.value) {
                case "true":
                    boolOut = true
                    break
                case "false":
                    boolOut = false
                    break
                default:
                    throw(`Illegal value: ${select.value}`)
            }

            maniplateUriValue(uri, boolOut)
            commitState()
        }

        console.log(`Listening on URI ${uri.string()}`)

        const event = 'MDCSelect:change'

        if (oldClosure !== null) {
            select.unlisten(event, oldClosure)
        }

        select.listen(event, newClosure);
        return newClosure
    }

    function newTextField(
        uri: JsonUri,
        inputType: HtmlInputType,
        labelText: string | null,
        currentValue: any,
    ): [HTMLElement, MDCTextField] {
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
        // TODO: validate input type
        inputElement.oninput = (_) => { onInputHook(uri, inputElement, inputType) }

        // TODO: add different values depending on the type?
        if (currentValue !== null) {
            inputElement.value = currentValue
        }

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

    function patchTextField(textFieldParent: HTMLElement, parentURI: JsonUri, inputType: HtmlInputType) {
        let inputInnerLabel = textFieldParent.querySelector('.mdc-floating-label')
        if (inputInnerLabel === null) {
            throw("Input inner label of textfield not found")
        }

        // Parse old URI
        const oldUri = new JsonUri(inputInnerLabel.id)
        // Create new URI
        let newUri = oldUri.clone()
        newUri.overrideStart(parentURI)

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
        inputElement.oninput = (_) => { onInputHook(newUri, inputElement, inputType) }
    }

    function patchDropDown(dropDownParent: HTMLElement, parentURI: JsonUri) {
        let selectRoot = dropDownParent.querySelector('.mdc-select')
        if (selectRoot === null) {
            // console.dir(dropDownParent)
            throw(`Inner dropdown not found of dropdown component`)
        }

        // Parse old URI
        const oldURI = new JsonUri(selectRoot.id)

        // Create new URI
        let newUri = oldURI.clone()
        newUri.overrideStart(parentURI)

        const newUriId = newUri.string()
        // Update id of label
        selectRoot.id = newUriId

        // Update on-input hook.

        for (let registered of boolDropDownStore) {
            if (registered.uri.equals(oldURI)) {
                const newListener = registerMDCSelectOnChange(registered.select, newUri.clone(), registered.listener)
                registered.listener = newListener
                registered.uri = newUri.clone()

                console.log(`Successfully patched dropdown with ID ${oldURI.string()} to ${newUriId}`)
                return
            }
        }

        throw(`Boolean input ${oldURI.string()} was not found in the list of registered inputs`)
    }

    function generateInputs(spec: ConfigSpec | null, topLevelLabel: string | null, currentData: any) {
        if (spec === null || dom === null) {
            return
        }

        const generatedTree = specToHtml(spec, topLevelLabel, new JsonUri(), currentData)

        // NOTE: Container must be cleared since reactive updates would cause repeated appensions which then causes
        // the configuration elements to be displayed multiple times.
        // TODO: Is it required / clean to somehow destroy the MDC components from JS?
        dom.innerHTML = ""
        dom.appendChild(generatedTree.html)
    }

    function drawUi(currentData: any) {
        generateInputs(spec, topLevelLabel, currentData)
    }

    // https://stackoverflow.com/questions/122102/what-is-the-most-efficient-way-to-deep-clone-an-object-in-javascript#122704
    function reactToNewInput(inp: any) {
        dataInternal = structuredClone(inp)
        drawUi(structuredClone(inp))
        commitState()
    }

    let loaded = false
    $: if(loaded) reactToNewInput(inputData)

    // TODO: is this `loaded` variable required?
    onMount(() => {
        loaded = true
    })
</script>

<div class="configurator">
    {#if spec !== null}
        <div class="configurator__dom" bind:this={dom}></div>
    {/if}
</div>

<style lang="scss">
    :global(.config-option__list__body), :global(.config-option__struct__fields) {
        list-style-type: none;

        & > :global(li:not(:last-child)) {
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
