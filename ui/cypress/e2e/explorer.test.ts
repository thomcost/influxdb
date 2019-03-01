import {Doc} from 'codemirror'
import {FROM, RANGE, MEAN} from '../../src/shared/constants/fluxFunctions'

interface HTMLElementCM extends HTMLElement {
  CodeMirror: {
    doc: CodeMirror.Doc
  }
}

type $CM = JQuery<HTMLElementCM>

describe('DataExplorer', () => {
  beforeEach(() => {
    cy.flush()

    cy.signin()

    cy.fixture('routes').then(({explorer}) => {
      cy.visit(explorer)
    })
  })

  describe('raw script editing', () => {
    beforeEach(() => {
      cy.getByTestID('switch-to-script-editor').click()
    })

    it('enables the submit button when a query is typed', () => {
      cy.getByTestID('time-machine-submit-button').should('be.disabled')

      cy.getByTestID('flux-editor').within(() => {
        cy.get('textarea').type('yo', {force: true})
        cy.getByTestID('time-machine-submit-button').should('not.be.disabled')
      })
    })

    it('disables submit when a query is deleted', () => {
      cy.getByTestID('time-machine--bottom').then(() => {
        cy.get('textarea').type('from(bucket: "foo")', {force: true})
        cy.getByTestID('time-machine-submit-button').should('not.be.disabled')
        cy.get('textarea').type('{selectall} {backspace}', {force: true})
      })

      cy.getByTestID('time-machine-submit-button').should('be.disabled')
    })

    it('can use the function selector to build a query', () => {
      cy.getByTestID('functions-toolbar-tab').click()

      cy.get<$CM>('.CodeMirror').then($cm => {
        const cm = $cm[0].CodeMirror
        cy.wrap(cm.doc).as('flux')
        expect(cm.doc.getValue()).to.eq('')
      })

      cy.getByTestID('flux-function from').click()

      cy.get<Doc>('@flux').then(doc => {
        const actual = doc.getValue()
        const expected = FROM.example

        cy.fluxEqual(actual, expected).should('be.true')
      })

      cy.getByTestID('flux-function range').click()

      cy.get<Doc>('@flux').then(doc => {
        const actual = doc.getValue()
        const expected = `${FROM.example}|>${RANGE.example}`

        cy.fluxEqual(actual, expected).should('be.true')
      })

      cy.getByTestID('flux-function mean').click()

      cy.get<Doc>('@flux').then(doc => {
        const actual = doc.getValue()
        const expected = `${FROM.example}|>${RANGE.example}|>${MEAN.example}`

        cy.fluxEqual(actual, expected).should('be.true')
      })
    })

    it('can filter aggregation functions by name from script editor mode', () => {
      cy.get('.input-field').type('covariance')
      cy.getByTestID('toolbar-function').should('have.length', 1)
    })

    it('can delete a second query', () => {
      cy.get('.time-machine-queries--new').click()
      cy.get('.query-tab').should('have.length', 2)
      cy.get('.query-tab--close')
        .first()
        .click()
      cy.get('.query-tab').should('have.length', 1)
    })

    it('can remove a second query using tab context menu', () => {
      cy.get('.query-tab').trigger('contextmenu')
      cy.getByTestID('right-click--remove-tab').should('have.class', 'disabled')

      cy.get('.time-machine-queries--new').click()
      cy.get('.query-tab').should('have.length', 2)

      cy.get('.query-tab')
        .first()
        .trigger('contextmenu')
      cy.getByTestID('right-click--remove-tab').click()

      cy.get('.query-tab').should('have.length', 1)
    })
  })

  describe('visualizations', () => {
    describe('empty states', () => {
      it('shows an error if a query is syntactically invalid', () => {
        cy.getByTestID('switch-to-script-editor').click()

        cy.getByTestID('time-machine--bottom').within(() => {
          cy.get('textarea').type('from(', {force: true})
          cy.getByTestID('time-machine-submit-button').click()
        })

        cy.getByTestID('empty-graph-message').within(() => {
          cy.contains('Error').should('exist')
        })
      })

      it('show an empty state for tag keys when the bucket is empty', () => {
        cy.getByTestID('empty-tag-keys').should('exist')
      })
    })
  })
})
