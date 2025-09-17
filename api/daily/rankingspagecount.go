/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package daily

import (
	"context"

	"github.com/pagefaultgames/rogueserver/repository"
)

// /daily/rankingpagecount - fetch daily ranking page count
func RankingPageCount(ctx context.Context, category int) (int, error) {
	pageCount, err := repository.Repos.Daily.FetchRankingPageCount(ctx, category)
	if err != nil {
		return pageCount, err
	}

	return pageCount, nil
}
