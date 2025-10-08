// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package jws

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/harp/pkg/sdk/value/signature"
)

// -----------------------------------------------------------------------------

var rsa2048PrivateJWK = []byte(`{
    "p": "_yLzpupxMheh6-VYntvlZPRFheEezWnr_7Q8dD73WrfXOpU296kw7dLaR4W8uJTscHGphW9qD4iuHCXQ8O7E4CGNk3gLMqnN7RS11WfRyidQ414SHot9DhozxKI02yYlU4nGJvCUvI14Q5E2Yc12BZYMH5oK6HPPrpUDGlEJrl0",
    "kty": "RSA",
    "q": "qh8Nc7PqKTfE5STnexTGJfb9D225cXjNFilzQpwyxCcAd7hYdvk2j6X2cNxKTFMZLQBfE90g7ItoEq2bZ3Oe0BixgI7gNVqqkOblKYiJ3je2K06Td1X2MEyC6o55XLpShjRXHK1bCUnmIdMnJ0iUhAcYsaOznELufRJqZgN0FZk",
    "d": "GCsQgKZ8JVRxtredwbTEwF02CUvsS9xv1GpYlcE_g4COF6ulxDRtyYLFJtNeSa4dT7pEHNM6y7bdMEVOVTaU6EWkyCIQTJ1NUdnS7qa5uJDkKhgYN87GaM0VJLj6Ks0Xp9O4Ljw1ASKMQcQVdDR9MZIBaN14cRGUF87BDQjR-M6v0CaGVOmMzF2iFjGmxWux8XJpZhIVUcWJPIhyGYyDkZcDOF6mAghfBwiAysVBgwoTrmY4rtST2uwTNhhTENvHk1n0dG9shpQFWsg7Ml27TSDCtEPp6VKxhIldOw1auV-AcVttCT8VrbVH6ENiDGKaTBptb8-mmWf0KgQtTOyqwQ",
    "e": "AQAB",
    "qi": "9dOJ8-lpm6dSu12pB2b9IN4u1yzbqMyrkVPnu06huZSFDqYhVS1m1pBB01tSiZjFmek0HmC7gqE0jOnfHN7myRN1gJMIayDNpBt2YiUX7tXO4mLpmHRpHccDkeJKY9b6KBiqq73fmCQLjsYiHhjAhuly6FhDcXMG1hTNmghFyKg",
    "dp": "i3id3Vc3vfxxRbVANsdapU9rWf5wijYnDseGbL2nFm59N8cuH0DcQIYNUR_oxLaUcfuQgiMfxZIgK774c1zzPtrYvA7ZgD6XFc1GgXyZkHrhmwlnuk7ic_0a45kQb-RwbzRXoB3QESr83WLaaaiZpgAabIQXBm-YzoUjBxXZxvk",
    "dq": "ny4THjI4ZnziZS6U0hvsFFX5D9ixbjWEYLpEOQ2CivubemopjJk_rkWyRIQ7mRMgOXVhgWqlhfAjX2bGRJRxA3I3lH0b2ZCOjKsfvi8eIamrcPZDEaSDiUnuLZ9Ge7dzxFNxN0oWJqjZBslWl3FYVfl157GvPAV8tdbna7DLi_E",
    "n": "qYwocG9HTEBtCp0yg0Z-Tfz_IxMryMZ_8O7fgK41gu8byv7HYY97iBlkuQpHfv9Ch8xePem8_3FPK9vKI8I8lZTT7f3t4LHYL_yfm9wOqy08NErRr6BeeW_NoLH686X3QZYBQmGIx0NlIEpLwAofZ9X9jr1WK5JPSa7M-DbHbgVLnex0iktmbvtamVejwJ4nCdoJPIZ1P9y8srWmlBj_Mf_kRPO86pgv557lhftMYbw_y2t_u7l6qs1LVGymrhmaK-3ZnT_xotInPx2LNFfd8MJTX-9cCgX7rlCS6-orkLsz0BZQg8weih5DICq9F0PGI6iBEPFHvlI-x9PcUzrWlQ"
}`)

var rsa3072PrivateJWK = []byte(`{
    "p": "uN30W1rjK5Pj__VFKPOi0J6e14eXAkA9DLvbQfVbym_b2KGgRMPtHIOlsz8GSn1KvE4docU9cfjUDSVeTJGcx4DEgvhAeXV15Rtpn7P1eUUSjr_HaSxeqqjveKccmbYGUvaMiN-mHPNQubBwxBsNO7RT-P_Xntxi58QxTXwHftwqE16ACZgZVcypaddRQxbvfWpDIymkv8OOSigRrIELIrnDls5g7z09bV27ZvdFdxDJF1jIOdgl18sKAmtuuYB_",
    "kty": "RSA",
    "q": "s63k6YU1oTUq-Ejzls7xdibrk2bMoVIRVZ29qXNjEbvE-1CgFTbMuZ3bHk6PGZ2_ss5hPIGyWr0dnWrmvgWHMvqpq4iRg-5iJeU9WeRnBppEbUrc-2MtmwFyzQELmqkgytBWZdpZI6x4cE7bGjSEX_UOuxQxZa0JtCr9iRelM5cU2W0gCTXDMEU2zCf8veKnd5DaxgoF8ziVrmHZqCQqNOukiBsRiYeYx9CNjRgNZqljkS2ngnr8FzZSQCvv_jgH",
    "d": "R_MSINKxb5RcA4WVOF1f2CkncvsGsTOsONib9_Cg8ieRggwKqGtDYe0zv1LZ4j1cvv4e_mpcL9TUKeAVKOVUgmlF25v3M9k7FZtizVoqO9wKXEFZeMMJ608iBFv5mcNmaDz2RqE8C3ldYFxCjBsKZAuodygPKFSysfGPpV12yl_uzIvj9-SW0Ow1rN_pFBTNHU5HfAW4I5vQW30QJO4lzJtQ_mHQLrGbNiDrGY-IEIJ8WR49dYFpR7cBJoJXxQkq14X0po6GG7RWwtfnA86ddfI_KTDW_Poq2xcAD6o9h1Y-Y2meBUYwb4_cMjf0rrFoA4CtCkxxubM6HHmv7xmvnLCBUBh_CG-7fhFkeRFBwml6tyRq1YCzqjHCrwO5z3PDL1Jsb7c4t9K2rOzOi7Yrv5YPj3jf2-pVnCWiWC1xd7b7TNOCNJXfx2fuDHHsKqVsrNYUFCiagcb4gfZP9mmoQXDe5W3TErLkfnpyb2hWsaR631yNHiiXKMH41fub0AeF",
    "e": "AQAB",
    "qi": "h5WO63K6wpvju-sDSqouGy6i0SdPLg5xzcBtmdci99kkHkSkb2Wx-WgApcQK5ZNk4esMHmb59-JZmL0a6l4xU1I7CNeKS1ywMxDMGDia1Ub5UyZaaBKtK8K6LAf35MwvJr60uS-T-arnf42tSeFwUUOuMm2fP8-JnFhsjHpla1m2D8IKzzoqyy2IohO8UP6YeUrdkfqYPU6TWQrteqI4d-jRUVNUR2O-lHe5g9tVr229MUIeQqOuP2G_xwL5e9l2",
    "dp": "H5_AbFxI3iyHZULE2POMl5l63K3SNE1e5C3CnS7my-OEvTMw4nKNvkH694XBgr6aKUpHoWPHVhbV14Xb8sc6BZrOgwUKqnxgdZfR4sL7LAqX2RmPeDth2lnZ1U7pApZz0H9inQt9NBifZ5R6ReHGyr7XWpIRhZh3xuP19yZPdfEdtYCF8J43P0kqdfOicBKdV1v1Li-ygidm0OK320Wb7Q3QGp0lUees-Wd-ZmfIV0zkyXyji9mg56BRDwbnww3r",
    "dq": "RCNc3OXEWbcE7ZUUswsVbH4D9ikUztSFeFdSdnEoj7AKWlfP7VuTtYxUbSDTiZ5V9SPoof55OYngMjM4_7Su-bkvILqwnDyJgd92LaG9JDbaXiv2s2p___lDpNSRpvweUGtDaGquRSYBom11D3H9BTq0dGOK_Ajr1iQ7c4LBBlhz8qQ8cuGtfJ-y0ScU5JmQcno_TLy-VucNdEztMqNMUjdOdML3GpWOba_8Dhah9l2kQYjzrjqfYNobxPQkld5_",
    "n": "gcDFKFIgEWXvugSrKgKF-x5EWHd8C9KzVvDMAl8BvtfS0InvHF3_tmeFoO7lz0L6WIo5jnChP1CmsG92q2hNUT3Jx044MuRF2Aws8sjVHvnhRkQExuF0C2tYOjNeDDjV1jjsGMO9cxuu8gOrYqr05g5Nv2Dl0ZrLIuu6_FLX0yHifMZm1Ri5Rymdio2XgOeD3VZc2Nzih1KCuYuGimj-obRll40x0H292fJ4JwQpHH-YDN4ANaRfRrTdF3YWRRyusUxNctr1i14JUkZkZN7tfZFe6Fl1RQ7e8Z4gTaGS73E3wF63galyHaXGHy0cKWfHvLXHNsPZEYUanPk4RoOpKYhM7elxw1Xo3RKFRyNKGrISzmqAghZoJdXBWwoZ16Lqz1eKjC_xvA1Hulx5y13Q1jTJtc6ksKrwJv4AMixqXkXffku8FSS4pD70v3mtUX8XD5mUZjAoBX38EF5ezqwsRReT04ONvLnQSwZn9jMjWDWNrfPKMpV1Kly68YGqMEt5"
}`)

var rsa4096PrivateJWK = []byte(`{
    "p": "1F-nnNrrx5qLjApeW0MFw5zDN45DilrE54jCuYG_5VQCgJ-b2ldySjRwAYOS-74yDVeYyr6VicJZJBQdExk4-w78xP1A_N_AgjINwLZI_-29s63qMPJKHym8AaaGKLa_eKZHbKzEmDgC3Xu0XRaPgI7MNvKTDnSaRaNnnefvRxkjaeDnnvOzGF9PXdM4IQ2iXrkRjdswkXwWnW4vLjhqHFqjHOt7GkKYHcp6zfcSgpZUsVPiR29O6cZA-PaLbDX6eGlGsqbLIDdLUdXD7GGc9YtP-galcdHeKrxtdCCPOS7EPhcVu3Weas8ZqTFyphWcUj927gj4UEyc6yZTI9M5pQ",
    "kty": "RSA",
    "q": "qrTYkllcUZgBe7TMI0IQrdRmjWvUKy4QLDJPmlqEPzDkFH7sgz-R5Ev1DlD6w7R47v-uUu3HZBEXrxUw3o2jHbmRAj02iUc3492eZveBNr8ElAXOko_9l2jx8SjfGFjrLB_b0r_Z3lfLZt2EsT8qRwWrDvNSI3f2HETtnCsjlcj_s2brFXcHgkAi2iMmyHAaiBwIGRZ8LmtB5gCWmqAOnCoRIr15QqUUnrSOriCWNUuLkLWh7ZmfLMXPj9HmYBgXzsqiEc0WUMQryVmHL_yrVL8CrNPrg2fwvkjPbkE_SCRzhGlH8hxm03Db2W7S3svZfOyYtucpNrOmM5e9uL-Yjw",
    "d": "E1d8SohIY5F3bNPBikL_Ttqcz-gU6AFqL6m0xJB-CIAGvwsjCEGcyaojGwvRaqFTvx1KhIF5x5ViA7C4rGfrENS76ti3uZVUclbZdcgpC-3zxm_jSFghmdJ4C4cyGClSvJh9WOhvDfNHm8iU24OtC_-CVxl203CudWaG5OubfwvUFExztTaHN_knfNjDDsMVNRJTi99BEW26YvsqUk60vlQ4loxkY2tIDqGmzurAIbkDzGZ0oAnHFyH2Y7ysh4HqU3XzHYCM-k96UzuIpcSxL6IMi-30LlUcvk7cBjGtkU6W5rpGm8QYfBIFPAK9iew_36kft_lDRaxZB86C92ya-5Av96Tgcq6EEPhUt4dNhKVFzBGQ2YvQGQyX__xdR_ykGNUoifVkwJywbjRxQ0XJ5tNGLwjnJtPikAW-MSnyRL23mhKI1VwCs1HXSGEJNlJp2R3Ib4MaZRPrqQ-5jbDwHtYWFwae-9e8FzDKrixPhvyJVbaTIRCaClgZdWeDAKSvElBkvTAjchlCam5SmlxsAbT_BqQB5temOYoQjOriHK5flB69u6dqP-9wGSCgJSzfG8a1Tv791q_NuZOdA3m4nA3gwAhxjDX-2J0vp7G-6Lkpw1Sw4qgL0qfw6Jg__QFp2X0QMManNKKUEK0-OhQvAYICPiZFCdmFb4QaSuEmx0k",
    "e": "AQAB",
    "qi": "VQYPsxi5IwWgYlKjUP_CvquMeQce8_gO5QxdvaaXPr2Zp_zZhdoHbcrN5R7D9cOqZhPARTo7yuFVrsCRKU-38Pcs_yN4zvy09SlI2kA4Vwi9ME5le2SXeGQgkey0acsIEUX4cYBSJ85fK5ZJTeLBBvacsdyAYAfQjvj224mhw_dyHJ6RepoXeyL7YE0VBI4duBzVx8jV4H0H65t-9fuu7woYQ-XmSElJPM3H265VJW-4YDLJ7QR5gpjm9bO0-Es14do-zHtbETPl77SBVjczILfp9TT2a-Kzgu7ckFCdG22tK8I8KxMdq8elMxT5oBmdaiMuyio1wdelWI-koN_auA",
    "dp": "SxylwIFhQNB1KIuGQcrboAqytNR5KNbfq3AgRtIQF8D1vZ411ix1fK89KhvqAWWMeFGR9asgYn-9XVhLzDRwhcMQPW_A628LvECNwqn1-aaRIJqTKMqY0prFJuRJsN6pq7dLrbEROzEcOk-FRCM1j-dxbMGwpz0wTw7zF9MvOrs2Xj0wTnSs1CLphCqvQGl9EvlrrvtmGx6DNR0CTNuhE_QumoziljAVcvvpIS2Qe2VGAQ3FCzTf-SioVCWGvDf_JVU-rvL1Bqjn18K-L716cRHbsHdnTnFGnWDVaWwWm8fhmoA5rHp2FBq5XbDidsFP73sIyjmb8XKXUUfVweb0DQ",
    "dq": "aVne5sMbhIepMq1r5r6ZCI6zE8heUp_E_2G1Wu9N-qmzuSpz0LRk115BSKqVeD5i_Czzat6wVYNu-HC9jjwfVPL3GUip5aL8TTay0Z0iM6VDsA69ZBpVSSJNXqX4uU_3I24t_izEysGoGD3R7ImtD9PhtAJayOT6EIxBkEXXWlOH6zIzFzY0fiAS4kkbYEw_M40JOmJ0FTnAm6_1QkxRd_NGTAhfU9AJn8CS5cfaq92Jrq9J1hce80TJSlsiMF_uIYNQ7meBxCqtU0BbfDQVkOGpRvwqtxhorSbGTaca5O0KFcfbzQHWO9vE8rXAkhuAh-aEgSy8dLP-eIzHW8Ny5Q",
    "n": "jZ2MOiHGR9bmwSRcflBMA9SEqZE56vWfpfawTJHvYhJW-Z5rHmufp2xUsLfa48r8cAcjtz2y3fhwJK-60eKG17JX56acvMqMEvhqS9Y10_v8ztOiFW7wPW9k0MJpvrfAypTVZarT4YWUJo72QAW8b-ZTpO53SdwwkL622oUArS7oaJPOSIloUbFViwhR5bEesUAcUeBlJlGBUlO4p_gxkKiYdfmez69VI-KGm80QhA0VziDrZLZxrtzk2ZAuVmKby7oPfFC5LppJlHWZtLSb0qzr819_w9y8-Sc4ZqqB33UfpA45UN9ufLYMeHaQOJCnXZRjcg6Qg8dVqXFdUgqmopsfl78YIF8swkuMe4Te1V8t4HLzCBVzAafo2THkHNgEILSPvEXfQq-En-ODb32HkEc3Zb_IyKq2aK67GvuAq2s_WlSYpOhz4rTmhez6yrBm52R9VSouSWGJhuRFLtBBvS613g5CykyTJlojXiZvrK68UCU7bZTa4zZmrJl3uNlizyhZr6qCwoSO6As0QKAOdVgcKUTNxv3h1NFmOHvMMIxAtQBdFvHaEjNjVvRkeS9waLuPySRihm5N3sBIv1NUKCSBU8tynMJhRUkf0Tp5EMeEBgQuqR-Cs5cgMcq_FD5s1ztT8f5ze9eCbt6Z9csnWaN2xO0-nPtMI7QzjgdSKys"
}`)

var p256PrivateJWK = []byte(`{
    "kty": "EC",
    "d": "sXCIy5HxtyG24MTl3hsgLDqi0dd33WAB_Rae1I_o2Is",
    "crv": "P-256",
    "x": "ykS0SN-EaFIVQUBC7norE9yYAN0ZFxSYYP6p0iofMxw",
    "y": "faQhXipqrhZeHIPFzJEYlxVvCdezZnJs2mKxnraO8_M"
}`)

var p384PrivateJWK = []byte(`{
    "kty": "EC",
    "d": "7YcsmkNxmZdzGyb46ZeDb2I1yr-ja1iw9gspGjq7UDqQ6a61h_ES8c4uU__adkFV",
    "crv": "P-384",
    "x": "dWLSo6PTkL1G68bzTwY3zzrL_QX-pwvP9HUPpQGeSFmj20EWOtfvXXKDrCR0jnJD",
    "y": "lFvTFechH_KmbOEvycryCHy23Cm1qekJYAtn7T_TELpm7zsY290NYlvqDKesGeXx"
}`)

var p521PrivateJWK = []byte(`{
    "kty": "EC",
    "d": "AIqIPpDjCGGwdG1usjkOkzovnv0SMiMgfLTn938E_gp4NBEyQVy4myOilDAEKrxPWw8f1u3FLKhGza-yxevMnfnr",
    "crv": "P-521",
    "x": "AVfi6aKylpZU334mETb2lNO5Ckpzp_L06WG4UQpiFxQMdxxKeldRJTxgt3FCYg5rXbUcKB2vm7Yq1Mxl3CHeBGQ8",
    "y": "AQQurRdp6oLjLbOTosM2cnu91dBL2YShDnqXbaUyFlGYoUJB6LPwwph9Uu0qHKCeK6QxZmHWxST2iky7ObEfM8GC"
}`)

var ed25519PrivateJWK = []byte(`{
    "kty": "OKP",
    "d": "ytOw6kKTTVJUKCnX5HgmhsGguNFQ18ECIS2C-ujJv-s",
    "crv": "Ed25519",
    "x": "K5i0d37-eRk8-EPwo2bpcmM-HGmzLiqRtWnk7oR3FCs"
}`)

func mustDecodeJWK(input []byte) *jose.JSONWebKey {
	var jwk jose.JSONWebKey
	if err := json.Unmarshal(input, &jwk); err != nil {
		panic(err)
	}

	return &jwk
}

// -----------------------------------------------------------------------------
func Test_jwsTransformer_To(t *testing.T) {
	type fields struct {
		key jose.SigningKey
	}
	type args struct {
		ctx   context.Context
		input []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "nil key",
			fields: fields{
				key: jose.SigningKey{
					Key: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "nil key",
			fields: fields{
				key: jose.SigningKey{
					Key: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid key",
			fields: fields{
				key: jose.SigningKey{
					Key:       &jose.JSONWebKey{},
					Algorithm: jose.RS256,
				},
			},
			args: args{
				ctx:   context.Background(),
				input: []byte("test"),
			},
			wantErr: true,
		},
		{
			name: "public key",
			fields: fields{
				key: jose.SigningKey{
					Key:       mustDecodeJWK(ed25519PrivateJWK).Public(),
					Algorithm: jose.RS256,
				},
			},
			args: args{
				ctx:   context.Background(),
				input: []byte("test"),
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid - HS256",
			fields: fields{
				key: jose.SigningKey{
					Key: &jose.JSONWebKey{
						Key: []byte("scye7sLSvuw9pB9bfkqZNoQ01CzjCtFhg64QcqQ60JU"),
					},
					Algorithm: jose.HS256,
				},
			},
			args: args{
				ctx:   signature.WithDetermisticSignature(context.Background(), true),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("eyJhbGciOiJIUzI1NiJ9.dGVzdA.YZ2zjgjYgXQA4kp3AWjt72XN6RUwxs2EAWpjTWTw2sA"),
		},
		{
			name: "valid - HS384",
			fields: fields{
				key: jose.SigningKey{
					Key: &jose.JSONWebKey{
						Key: []byte("ZO-sssUGsRpYzgTLHH7SHL410d7S0ekSaJudOun8k3s-kM_9GUqr3BCpbZfAK1rk"),
					},
					Algorithm: jose.HS384,
				},
			},
			args: args{
				ctx:   signature.WithDetermisticSignature(context.Background(), true),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("eyJhbGciOiJIUzM4NCJ9.dGVzdA.s2vnkZWuRntbEHvWZL5Da22n5tpfqAn9In6Nc59oXkThtjoHO8YV_xFoBszoNbff"),
		},
		{
			name: "valid - HS512",
			fields: fields{
				key: jose.SigningKey{
					Key: &jose.JSONWebKey{
						Key: []byte("Wgl6uSlvowSnVQhR0bkJ3uun-IJiJn0o3CfwOcH0IgTVHiBBgVSF4z2KVWW6RATGWjx5zjCk6FUtq9Jx-eoXvw"),
					},
					Algorithm: jose.HS512,
				},
			},
			args: args{
				ctx:   signature.WithDetermisticSignature(context.Background(), true),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("eyJhbGciOiJIUzUxMiJ9.dGVzdA.sRArznPrS3SDMLdOplA1gW9sDvPK_bS78S8LjqKHdsRRWGmcU673RM8W20C66RLJqST8g1lK2rlhauBbys-YVQ"),
		},
		{
			name: "valid - RS256",
			fields: fields{
				key: jose.SigningKey{
					Key:       mustDecodeJWK(rsa2048PrivateJWK),
					Algorithm: jose.RS256,
				},
			},
			args: args{
				ctx:   signature.WithDetermisticSignature(context.Background(), true),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("eyJhbGciOiJSUzI1NiJ9.dGVzdA.M24oz_l9RSn9qt4aVoEVXg4EvcgtanpikhMs1BkL3JbCoBi9-M9EZdMUZk4_yLUdj-NmDO4O20V_JjxtlyXMm45AGTTghM_f3aBo5eqBWVkJuyl8EqJM2HeowmoyY7ZxlsoJwEA3VSA68CUSaukjd33zOF4SGPea5aDZClTjdC6Y-OvTzlDX1l6q99fPDrRJO-ih4gdsWLs-EYKaI-nTPzvtGcjOAfchTChT-kH0DKjANf35eMbceTNlZODlQfWw0vB2Zqeu4U8SkXpk6oA3S25COpcXjx-k8sxZbrLq57M0jppgRIrMxq-JvORZ716U3B2cRAVmP1SwgqcBnsLLmA"),
		},
		{
			name: "valid - RS384",
			fields: fields{
				key: jose.SigningKey{
					Key:       mustDecodeJWK(rsa3072PrivateJWK),
					Algorithm: jose.RS384,
				},
			},
			args: args{
				ctx:   signature.WithDetermisticSignature(context.Background(), true),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("eyJhbGciOiJSUzM4NCJ9.dGVzdA.ML2P3gs_gC3R1jUflAggcQER7fe1OcENO_Rrei0PgKfvvfyE1x-5aTzoMZnIHT_5sxBdxa7wxssy6IYEZzE7vSCLxoJdlWgRaRfZo6LrEf8B6Q80kG7Oc6vBwEetzCLjJD_IBVYLOMqkOvxSaBdOB28NDFVBifvqA6-M1VHHN9koUerwwCMshJmH7dsjC083JsQeui2ThVedhYxOHb4K5IkCiYOJBnNF15E9qxIrCtZQQK5hAvk5-BEzcKrT6BRrUo1L_fidgE9tb2mIqDeWsap64DdIIMHJ9sxfcXg1c0CZ_Rocd3lVrpeCkNa2s6sY8HILOYG0Vz1EI1-rA_r_TAD3rMbS-l2ENZMbUKus7zbN1bzLQMR-D8KVB7uUnpFsOJqX47xl1D_QCBuOI5tHKCx0707U29BAe0Mysp2jFx1qskoL_JTl9ObVYV_JjJIAmxhxZ7TxQaMzfOqPpEzgoTHscSfyXd05_oUTSxISmxQf6S7lpUQ5tRIVT2qumiGL"),
		},
		{
			name: "valid - RS512",
			fields: fields{
				key: jose.SigningKey{
					Key:       mustDecodeJWK(rsa4096PrivateJWK),
					Algorithm: jose.RS512,
				},
			},
			args: args{
				ctx:   signature.WithDetermisticSignature(context.Background(), true),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("eyJhbGciOiJSUzUxMiJ9.dGVzdA.e201ath4i-njH626lFFadhOCrk0XJVlwhFxIkbaweTl6FlFaO9x178zLmYjVNkiC-ooVzFiYegHw0lqGHGX0CgMnLy97_6GJfnwLF6Nhmi7jFo2FtSCJ71-d7i6ObAJkWT_4PFzIemPc31xa0FGzKo8uG3Xvmug1Pz94H2V_4Hsc14M32vBv0vd5daf581YQTln_CTRajZEd46s6rTYA6PEFG9C4-N-3jiqvPd3aga8ZwX6S7nTYuQMnbRnX33g5ZEIjJNpJT0QInbbPmM0YmGzSYfwyGuso1bhRueqri0PnvoMlVe7EK_WnjH3-MGr1RcHDAbl4I9mqZBL_danAJqhcqKM6AaRVrcNaOiO5R-KNSMskEnP8JvLE0PXOlpmv886uOjEhrWV3KG2js1E908xdDZE5kZ2Dhvv2PZxGDtzYqenTGWMff6t3R4Md8xM7LloB3xjopqOKzqJeb3cdPnKG9g4hrYelyKUzAo3vV_fK9ZuR2FTLdLrzp2pdSDRv3CffKCLEkvqS7COTQBwx-NHPo3Zhq_X8CjldwmdvF-BcEK2tc8C2vRlDrwQvbIWNfUS9lAHbGMq6UbfTSnXD9nzJpn9bMGe-q2Q28kfcPB6IrTQXYrHo97M-3rXc5rhbpyEcheZCXbPWTtAOmnCV1c_py4xDyR9hh384fok9Yts"),
		},
		{
			name: "valid - EdDSA",
			fields: fields{
				key: jose.SigningKey{
					Key:       mustDecodeJWK(ed25519PrivateJWK),
					Algorithm: jose.EdDSA,
				},
			},
			args: args{
				ctx:   signature.WithDetermisticSignature(context.Background(), true),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("eyJhbGciOiJFZERTQSJ9.dGVzdA.UHM26VhesCXny588L0ou6Hj8xdyB5NnyYg_vQYPYMct7LIjEVf7_E6EeYz2wzvNdKoxmf5j8dpbjPzGg_pDzDA"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &jwsTransformer{
				key: tt.fields.key,
			}
			got, err := d.To(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("jwsTransformer.To() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jwsTransformer.To() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func Test_jwsTransformer_Roundtrip(t *testing.T) {
	testcases := []struct {
		name               string
		privateKey         *jose.JSONWebKey
		signatureAlgorithm jose.SignatureAlgorithm
	}{
		{
			name:               "ed25519",
			privateKey:         mustDecodeJWK(ed25519PrivateJWK),
			signatureAlgorithm: jose.EdDSA,
		},
		{
			name:               "p256",
			privateKey:         mustDecodeJWK(p256PrivateJWK),
			signatureAlgorithm: jose.ES256,
		},
		{
			name:               "p384",
			privateKey:         mustDecodeJWK(p384PrivateJWK),
			signatureAlgorithm: jose.ES384,
		},
		{
			name:               "p521",
			privateKey:         mustDecodeJWK(p521PrivateJWK),
			signatureAlgorithm: jose.ES512,
		},
		{
			name:               "rsa2048 - rs256",
			privateKey:         mustDecodeJWK(rsa2048PrivateJWK),
			signatureAlgorithm: jose.RS256,
		},
		{
			name:               "rsa3072 - rs384",
			privateKey:         mustDecodeJWK(rsa3072PrivateJWK),
			signatureAlgorithm: jose.RS384,
		},
		{
			name:               "rsa4096 - rs512",
			privateKey:         mustDecodeJWK(rsa4096PrivateJWK),
			signatureAlgorithm: jose.RS512,
		},
		{
			name:               "rsa2048 - ps256",
			privateKey:         mustDecodeJWK(rsa2048PrivateJWK),
			signatureAlgorithm: jose.PS256,
		},
		{
			name:               "rsa3072 - ps384",
			privateKey:         mustDecodeJWK(rsa3072PrivateJWK),
			signatureAlgorithm: jose.PS384,
		},
		{
			name:               "rsa4096 - ps512",
			privateKey:         mustDecodeJWK(rsa4096PrivateJWK),
			signatureAlgorithm: jose.PS512,
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			signer := &jwsTransformer{
				key: jose.SigningKey{
					Algorithm: tt.signatureAlgorithm,
					Key:       tt.privateKey,
				},
			}

			verifier := &jwsTransformer{
				key: jose.SigningKey{
					Algorithm: tt.signatureAlgorithm,
					Key:       tt.privateKey.Public(),
				},
			}

			// Prepare context
			ctx := context.Background()
			input := []byte("test")

			signed, err := signer.To(ctx, input)
			assert.NoError(t, err)

			payload, err := verifier.From(ctx, signed)
			assert.NoError(t, err)

			assert.Equal(t, input, payload)
		})
	}
}
