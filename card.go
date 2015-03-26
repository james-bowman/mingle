package mingle

import (
	"encoding/xml"
)

type CardType struct {
	XMLName xml.Name `xml:"card_type"`
	Name string `xml:"name"`
}

type Card struct {
	XMLName  xml.Name	`xml:"card"`
	Name string `xml:"name"`
	Description string `xml:"description"`
	Type CardType `xml:"card_type"`
	Id string `xml:"id"`
	Number string `xml:"number"`
	Vesion int `xml:"version"`
	
	/* name: String.
description: String; the HTML that Mingle renders for the card description.
card type: Resource; name of the card type for each card; string.
id: Integer; read only, system assigned unique identifier for a card.
number: Integer; read only, unique identifier of a card within a project - Use this for both GET and PUT.
project: Resource; name and identifier of a project a card belongs to; both strings.
version: Integer; read only, current card version. You can specify the version to get history version of the card.
project_card_rank: Integer; read only, the rank of the card in a project.
created_on: Datetime; read only, date and time of creating card.
modified_on: Datetime; read only, date and time of last modification.
modified_by: Resource; name and login id of user who is the last to modify the card; both String, read only.
created_by: Resource; name and login id of user who created the card; both String, read only.
properties: Array; property: Resource; name and a current value for each card property defined for current card's card type are listed; Data type will depend on the property while property name is always String. The property also includes attributes about the property type_description and whether or not it is hidden.
tags: String; read only, comma-delimited list of tags associated with the card.
rendered_description: Resource; Link to rendered card description as HTML.*/
}
