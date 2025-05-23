// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {backend} from '../models';

export function CreateMovie(arg1:string,arg2:string,arg3:Record<string, string>):Promise<backend.Movie>;

export function DeleteMovie(arg1:number):Promise<void>;

export function GetMovieByID(arg1:number):Promise<backend.Movie>;

export function ListMovies(arg1:string,arg2:backend.Pagination):Promise<backend.ListMoviesResponse>;

export function UpdateMovie(arg1:backend.Movie):Promise<void>;
