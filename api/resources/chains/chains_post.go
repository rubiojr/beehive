/*
 *    Copyright (C) 2017 Christian Muehlhaeuser
 *
 *    This program is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU Affero General Public License as published
 *    by the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    This program is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU Affero General Public License for more details.
 *
 *    You should have received a copy of the GNU Affero General Public License
 *    along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *    Authors:
 *      Christian Muehlhaeuser <muesli@gmail.com>
 */

package chains

import (
	"github.com/emicklei/go-restful"
	"github.com/muesli/beehive/bees"
	"github.com/muesli/beehive/cfg"
	"github.com/muesli/smolder"
)

// ChainPostStruct holds all values of an incoming POST request
type ChainPostStruct struct {
	Chain struct {
		Name        string     `json:"name"`
		Description string     `json:"description"`
		Event       bees.Event `json:"event"`
		Filters     []string   `json:"filters"`
		Actions     []string   `json:"actions"`
	} `json:"chain"`
}

// PostAuthRequired returns true because all requests need authentication
func (r *ChainResource) PostAuthRequired() bool {
	return false
}

// PostDoc returns the description of this API endpoint
func (r *ChainResource) PostDoc() string {
	return "create a new chain"
}

// PostParams returns the parameters supported by this API endpoint
func (r *ChainResource) PostParams() []*restful.Parameter {
	return nil
}

// Post processes an incoming POST (create) request
func (r *ChainResource) Post(context smolder.APIContext, data interface{}, request *restful.Request, response *restful.Response) {
	resp := ChainResponse{}
	resp.Init(context)

	pps := data.(*ChainPostStruct)
	dupe := bees.GetChain(pps.Chain.Name)
	if dupe != nil {
		smolder.ErrorResponseHandler(request, response, smolder.NewErrorResponse(
			422, // Go 1.7+: http.StatusUnprocessableEntity,
			false,
			"A Chain with that name exists already",
			"ChainResource POST"))
		return
	}

	chain := bees.Chain{
		Name:        pps.Chain.Name,
		Description: pps.Chain.Description,
		Event:       &pps.Chain.Event,
		Actions:     pps.Chain.Actions,
		Filters:     pps.Chain.Filters,
	}
	chains := append(bees.GetChains(), chain)
	bees.SetChains(chains)

	cfg.SaveCurrentConfig()

	resp.AddChain(chain)
	resp.Send(response)
}
